package lyrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// ErrNoLyrics means the provider answered "nothing here", not "provider broken".
var ErrNoLyrics = errors.New("lyrics not found")

// RemoteLyrics is one LRCLIB record.
type RemoteLyrics struct {
	ID           int64   `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
	HasWordSync  bool    `json:"hasWordSync"`
}

// LRCLIB talks to the public LRCLIB API.
//
// Usage policy: one lookup per track, results cached in SQLite. The API asks
// for a descriptive User-Agent and no bulk scraping, which is what this client
// does.
type LRCLIB struct {
	http    *http.Client
	baseURL string
	ua      string
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
	}
}

// WithBaseURL overrides the endpoint for tests.
func (l *LRCLIB) WithBaseURL(base string) *LRCLIB {
	l.baseURL = strings.TrimRight(base, "/")
	return l
}

// Get asks for the exact match: title, artist and duration.
func (l *LRCLIB) Get(ctx context.Context, trackName, artistName string, durationSec int) (*RemoteLyrics, error) {
	q := url.Values{}
	q.Set("track_name", trackName)
	q.Set("artist_name", artistName)
	if durationSec > 0 {
		q.Set("duration", strconv.Itoa(durationSec))
	}
	return l.fetch(ctx, "/get?"+q.Encode())
}

// Search runs a free-text query.
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

// Lookup finds the best candidate for a track: exact match first, then a search
// filtered by artist and duration tolerance.
func (l *LRCLIB) Lookup(ctx context.Context, track domain.Track) (*RemoteLyrics, error) {
	if strings.TrimSpace(track.Title) == "" {
		return nil, ErrNoLyrics
	}
	durationSec := int(track.DurationMs / 1000)
	exact, err := l.Get(ctx, track.Title, track.Artist, durationSec)
	if err == nil && hasContent(exact) {
		return exact, nil
	}
	if err != nil && !errors.Is(err, ErrNoLyrics) {
		// A hard failure (network, 5xx) should be reported, not masked by the
		// fallback search below.
		return nil, err
	}

	query := strings.TrimSpace(track.Title + " " + track.Artist)
	results, err := l.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	return pickBest(results, track), nil
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
		if track.Artist != "" && strings.Contains(strings.ToLower(r.ArtistName), strings.ToLower(track.Artist)) {
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
	return r != nil && !r.Instrumental && (r.SyncedLyrics != "" || r.PlainLyrics != "")
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

func (l *LRCLIB) request(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build LRCLIB request: %w", err)
	}
	req.Header.Set("User-Agent", l.ua)
	req.Header.Set("Accept", "application/json")

	resp, err := l.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LRCLIB request failed: %w", err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNoLyrics
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, fmt.Errorf("LRCLIB rate limited the request; try again later")
	case resp.StatusCode >= 500:
		return nil, fmt.Errorf("LRCLIB is unavailable (status %d)", resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("LRCLIB answered %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("read LRCLIB response: %w", err)
	}
	return body, nil
}
