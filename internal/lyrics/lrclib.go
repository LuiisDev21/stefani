package lyrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// ErrNoLyrics means the provider answered "nothing here", not "provider broken".
var ErrNoLyrics = errors.New("lyrics not found")

// RemoteLyrics is one LRCLIB record (v1 shape).
type RemoteLyrics struct {
	ID           int64   `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  *string `json:"plainLyrics"`
	SyncedLyrics *string `json:"syncedLyrics"`
	HasWordSync  bool    `json:"hasWordSync"`
}

// SearchSummary is the lightweight v2 search record: metadata and availability
// flags, without lyrics payloads.
type SearchSummary struct {
	ID              int64   `json:"id"`
	TrackName       string  `json:"trackName"`
	ArtistName      string  `json:"artistName"`
	AlbumName       string  `json:"albumName"`
	Duration        float64 `json:"duration"`
	Instrumental    bool    `json:"instrumental"`
	HasWordSync     bool    `json:"hasWordSync"`
	HasPlainLyrics  bool    `json:"hasPlainLyrics"`
	HasSyncedLyrics bool    `json:"hasSyncedLyrics"`
	HasLyricsfile   bool    `json:"hasLyricsfile"`
}

// LRCLIB talks to the public LRCLIB API.
//
// Contract notes, from https://lrclib.net/docs:
//   - /api/get requires track_name AND artist_name; a missing artist is a 400
//     returned by the framework as plain text, not JSON.
//   - duration is the strongest matching signal (±2 s tolerance).
//   - /api/v2/search returns metadata only; the lyrics are fetched per record
//     with /api/get/{id}. That keeps one lookup to one small request plus the
//     selected payload.
//   - 429 and 503 carry Retry-After and must be honored.
type LRCLIB struct {
	http    *http.Client
	baseURL string
	ua      string
	sleep   func(time.Duration)
	now     func() time.Time
}

// NewLRCLIB builds the client.
func NewLRCLIB(timeout time.Duration) *LRCLIB {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &LRCLIB{
		http:    &http.Client{Timeout: timeout},
		baseURL: "https://lrclib.net/api",
		ua:      "STEFANI/0.1 (desktop lyrics client; +https://github.com/LuiisDev21/stefani)",
		sleep:   time.Sleep,
		now:     time.Now,
	}
}

// WithBaseURL overrides the endpoint for tests.
func (l *LRCLIB) WithBaseURL(base string) *LRCLIB {
	l.baseURL = strings.TrimRight(base, "/")
	return l
}

// Get asks for the exact match: title, artist and (when known) duration.
// Both names are required by the endpoint, so it is only called with both.
func (l *LRCLIB) Get(ctx context.Context, trackName, artistName string, durationSec int) (*RemoteLyrics, error) {
	if strings.TrimSpace(trackName) == "" || strings.TrimSpace(artistName) == "" {
		return nil, ErrNoLyrics
	}
	q := url.Values{}
	q.Set("track_name", trackName)
	q.Set("artist_name", artistName)
	if durationSec > 0 && durationSec <= 3600 {
		q.Set("duration", strconv.Itoa(durationSec))
	}
	return l.fetch(ctx, "/get?"+q.Encode())
}

// GetByID fetches one record by its stable LRCLIB id.
func (l *LRCLIB) GetByID(ctx context.Context, id int64) (*RemoteLyrics, error) {
	if id <= 0 {
		return nil, ErrNoLyrics
	}
	return l.fetch(ctx, fmt.Sprintf("/get/%d", id))
}

// Search runs a free-text query against the v1 endpoint, which includes lyrics.
// Kept for the fallback path and for callers that want the text in one call.
func (l *LRCLIB) Search(ctx context.Context, query string) ([]RemoteLyrics, error) {
	q := url.Values{}
	q.Set("q", query)
	body, err := l.request(ctx, "/search?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var out []RemoteLyrics
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode LRCLIB search: %w", err)
	}
	return out, nil
}

// SearchSummaries runs the lightweight v2 search: metadata plus availability.
func (l *LRCLIB) SearchSummaries(ctx context.Context, query, trackName, artistName string) ([]SearchSummary, error) {
	q := url.Values{}
	switch {
	case strings.TrimSpace(query) != "":
		q.Set("q", strings.TrimSpace(query))
	case strings.TrimSpace(trackName) != "":
		q.Set("track_name", strings.TrimSpace(trackName))
		if name := strings.TrimSpace(artistName); name != "" {
			q.Set("artist_name", name)
		}
	default:
		return nil, nil
	}
	body, err := l.request(ctx, "/v2/search?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var out []SearchSummary
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode LRCLIB v2 search: %w", err)
	}
	return out, nil
}

// Lookup finds the best candidate for a track: exact match first, then a search
// ranked by title, artist and duration.
func (l *LRCLIB) Lookup(ctx context.Context, track domain.Track) (*RemoteLyrics, error) {
	// YouTube titles carry "(Official Video)" and channels carry " - Topic";
	// both break matching, so the query uses the cleaned pair.
	title := cleanTitle(track.Title)
	artist := cleanArtist(track.Artist)
	if title == "" {
		return nil, ErrNoLyrics
	}

	target := track
	target.Title = title
	target.Artist = artist

	// The exact endpoint requires both names and is the highest-precision path
	// when the duration is known.
	if artist != "" {
		exact, err := l.Get(ctx, title, artist, int(track.DurationMs/1000))
		if err == nil && hasContent(exact) {
			return exact, nil
		}
	}

	// Search first, then fetch only the chosen record.
	summary, err := l.bestSummary(ctx, title, artist, target)
	if err != nil && !errors.Is(err, ErrNoLyrics) {
		return nil, err
	}
	if summary != nil {
		doc, err := l.GetByID(ctx, summary.ID)
		if err == nil && hasContent(doc) {
			return doc, nil
		}
		if err != nil && !errors.Is(err, ErrNoLyrics) {
			return nil, err
		}
	}

	// v2 is beta: if it is unavailable, fall back to the v1 search, which
	// carries the lyrics inline.
	results, err := l.searchV1(ctx, title, artist)
	if err != nil {
		return nil, err
	}
	if best := pickBest(results, target); best != nil {
		return best, nil
	}
	return nil, ErrNoLyrics
}

func (l *LRCLIB) bestSummary(ctx context.Context, title, artist string, target domain.Track) (*SearchSummary, error) {
	summaries, err := l.SearchSummaries(ctx, "", title, artist)
	if err != nil {
		if errors.Is(err, ErrNoLyrics) {
			return nil, nil
		}
		return nil, err
	}
	if best := pickBestSummary(summaries, target); best != nil {
		return best, nil
	}
	// A combined query is stricter; when the title alone returns nothing useful,
	// widen to the free-text form the documentation prefers.
	if artist != "" {
		summaries, err = l.SearchSummaries(ctx, title+" "+artist, "", "")
		if err != nil && !errors.Is(err, ErrNoLyrics) {
			return nil, err
		}
		return pickBestSummary(summaries, target), nil
	}
	return nil, nil
}

func (l *LRCLIB) searchV1(ctx context.Context, title, artist string) ([]RemoteLyrics, error) {
	queries := []string{title}
	if artist != "" {
		queries = append(queries, title+" "+artist)
	}
	var lastErr error
	for _, query := range queries {
		results, err := l.Search(ctx, query)
		if err != nil {
			if errors.Is(err, ErrNoLyrics) {
				continue
			}
			lastErr = err
			continue
		}
		return results, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, nil
}

func pickBestSummary(results []SearchSummary, track domain.Track) *SearchSummary {
	var best *SearchSummary
	bestScore := 0.0
	targetSec := track.DurationMs / 1000
	for i := range results {
		r := &results[i]
		if r.Instrumental || (!r.HasSyncedLyrics && !r.HasPlainLyrics && !r.HasLyricsfile) {
			continue
		}
		score := 1.0
		if strings.EqualFold(strings.TrimSpace(r.TrackName), strings.TrimSpace(track.Title)) {
			score += 2
		}
		if artist := strings.TrimSpace(track.Artist); artist != "" &&
			strings.Contains(strings.ToLower(r.ArtistName), strings.ToLower(artist)) {
			score += 2
		}
		if targetSec > 0 {
			diff := r.Duration - float64(targetSec)
			if diff < 0 {
				diff = -diff
			}
			switch {
			case diff <= 2:
				score += 2
			case diff <= 5:
				score += 1
			case diff > 12:
				score -= 2
			}
		}
		if r.HasWordSync {
			score += 0.5
		}
		if r.HasSyncedLyrics {
			score += 0.5
		}
		if score > bestScore {
			bestScore = score
			best = r
		}
	}
	return best
}

func pickBest(results []RemoteLyrics, track domain.Track) *RemoteLyrics {
	var best *RemoteLyrics
	bestScore := 0.0
	targetSec := track.DurationMs / 1000
	for i := range results {
		r := &results[i]
		if !hasContent(r) {
			continue
		}
		score := 1.0
		if strings.EqualFold(strings.TrimSpace(r.TrackName), strings.TrimSpace(track.Title)) {
			score += 2
		}
		if artist := strings.TrimSpace(track.Artist); artist != "" &&
			strings.Contains(strings.ToLower(r.ArtistName), strings.ToLower(artist)) {
			score += 1
		}
		if targetSec > 0 {
			diff := r.Duration - float64(targetSec)
			if diff < 0 {
				diff = -diff
			}
			switch {
			case diff <= 2:
				score += 2
			case diff <= 5:
				score += 1
			case diff > 12:
				score -= 2
			}
		}
		if r.HasWordSync {
			score += 0.5
		}
		if score > bestScore {
			bestScore = score
			best = r
		}
	}
	return best
}

func hasContent(r *RemoteLyrics) bool {
	return r != nil && !r.Instrumental && (stringValue(r.SyncedLyrics) != "" || stringValue(r.PlainLyrics) != "")
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (l *LRCLIB) fetch(ctx context.Context, path string) (*RemoteLyrics, error) {
	body, err := l.request(ctx, path)
	if err != nil {
		return nil, err
	}
	var out RemoteLyrics
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode LRCLIB response: %w", err)
	}
	if !hasContent(&out) {
		return nil, ErrNoLyrics
	}
	return &out, nil
}

// request performs one GET with the documented retry policy: honor Retry-After
// on 429/503, retry twice, and never block the UI for more than a few seconds.
func (l *LRCLIB) request(ctx context.Context, path string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		body, retryAfter, err := l.once(ctx, path)
		if err == nil {
			return body, nil
		}
		if retryAfter <= 0 {
			return nil, err
		}
		if retryAfter > 5*time.Second {
			return nil, fmt.Errorf("%w (retry in %s)", err, retryAfter.Round(time.Second))
		}
		lastErr = err
		// Jitter keeps parallel lookups from retrying in a synchronized wave.
		jitter := time.Duration(rand.Int63n(int64(retryAfter/5) + 1))
		if sleep := l.sleep; sleep != nil {
			sleep(retryAfter + jitter)
		}
	}
	return nil, lastErr
}

// once returns the body, or a wait duration when the server asked to retry.
func (l *LRCLIB) once(ctx context.Context, path string) ([]byte, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.baseURL+path, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build LRCLIB request: %w", err)
	}
	req.Header.Set("User-Agent", l.ua)
	req.Header.Set("Accept", "application/json")

	resp, err := l.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("LRCLIB request failed: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode == http.StatusNotFound:
		return nil, 0, ErrNoLyrics
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable:
		wait := parseRetryAfter(resp.Header.Get("Retry-After"), l.now())
		if wait <= 0 {
			wait = 500 * time.Millisecond
		}
		return nil, wait, fmt.Errorf("LRCLIB is busy (status %d)", resp.StatusCode)
	case resp.StatusCode == http.StatusBadRequest:
		// A rejected query (for example /get without an artist) is "nothing to
		// look up", not an outage: the caller must be able to fall back.
		return nil, 0, fmt.Errorf("%w: LRCLIB rejected the query", ErrNoLyrics)
	case resp.StatusCode >= 500:
		return nil, 0, fmt.Errorf("LRCLIB is unavailable (status %d)", resp.StatusCode)
	default:
		return nil, 0, fmt.Errorf("LRCLIB answered %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, 0, fmt.Errorf("read LRCLIB response: %w", err)
	}
	return body, 0, nil
}

// parseRetryAfter accepts seconds or an HTTP date, as the header allows both.
func parseRetryAfter(raw string, now time.Time) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		if seconds < 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(raw); err == nil {
		if d := when.Sub(now); d > 0 {
			return d
		}
	}
	return 0
}

// noiseSuffix matches the video-service decoration YouTube adds to a title:
// "(Official Video)", "[Lyrics]", "(Remastered 2011)", "(Live at Wembley)".
var noiseSuffix = regexp.MustCompile(`(?i)\s*[\[(]\s*(official[^\])]*|lyrics?|lyric\s*video|video|audio|visuali[sz]er|remaster(ed)?[^\])]*|live[^\])]*|hd|hq|4k|explicit)\s*[\])]`)

// cleanTitle removes the decoration without touching the words of the title.
func cleanTitle(raw string) string {
	title := strings.TrimSpace(raw)
	title = noiseSuffix.ReplaceAllString(title, "")
	title = strings.Join(strings.Fields(title), " ")
	return strings.TrimSpace(title)
}

// cleanArtist removes the channel suffixes YouTube appends to an artist.
func cleanArtist(raw string) string {
	artist := strings.TrimSpace(raw)
	for _, suffix := range []string{" - Topic", " – Topic", " - topic"} {
		artist = strings.TrimSuffix(artist, suffix)
	}
	return strings.TrimSpace(artist)
}
