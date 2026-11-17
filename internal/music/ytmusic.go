package music

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// YTMusic is an experimental catalog client for YouTube Music.
//
// Scope, deliberately: it reads public search metadata through InnerTube, the
// same endpoint the web client uses. It does not extract audio streams, does
// not touch DRM, does not use cookies or an account, and does not claim that a
// result can be played inside STEFANI. Playback of these entries stays in the
// official player, opened by the user.
type YTMusic struct {
	http    *http.Client
	baseURL string
	webURL  string
	browse  string

	mu            sync.Mutex
	apiKey        string
	clientVersion string
}

// NewYTMusic builds the client with a bounded HTTP timeout.
func NewYTMusic(timeout time.Duration) *YTMusic {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &YTMusic{
		http:    &http.Client{Timeout: timeout},
		baseURL: "https://music.youtube.com/youtubei/v1/search",
		webURL:  "https://music.youtube.com/",
	}
}

// WithBaseURL overrides the endpoints, for tests against a fake server.
func (y *YTMusic) WithBaseURL(searchURL, webURL string) *YTMusic {
	y.baseURL = searchURL
	y.webURL = webURL
	return y
}

func (y *YTMusic) Name() string { return "YouTube Music (experimental)" }

var (
	apiKeyRe        = regexp.MustCompile(`"INNERTUBE_API_KEY":"([^"]+)"`)
	clientVersionRe = regexp.MustCompile(`"INNERTUBE_CLIENT_VERSION":"([^"]+)"`)
)

// credentials fetches and caches the public web client configuration. The key
// rotates, so it is never hard-coded: a failed request refreshes it once.
func (y *YTMusic) credentials(ctx context.Context) (apiKey, version string, err error) {
	y.mu.Lock()
	defer y.mu.Unlock()
	if y.apiKey != "" {
		return y.apiKey, y.clientVersion, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, y.webURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("build credential request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := y.http.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("%w: music.youtube.com answered %d", ErrUnavailable, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", "", fmt.Errorf("%w: read page: %v", ErrUnavailable, err)
	}
	key := apiKeyRe.FindSubmatch(body)
	if key == nil {
		return "", "", fmt.Errorf("%w: InnerTube key not found in page", ErrUnavailable)
	}
	version = "1.20250101.01.00"
	if m := clientVersionRe.FindSubmatch(body); m != nil {
		version = string(m[1])
	}
	y.apiKey = string(key[1])
	y.clientVersion = version
	return y.apiKey, y.clientVersion, nil
}

func (y *YTMusic) invalidateCredentials() {
	y.mu.Lock()
	y.apiKey = ""
	y.mu.Unlock()
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// Search returns metadata hits. Errors are wrapped with ErrUnavailable when the
// upstream contract breaks, so the caller can degrade instead of crashing.
func (y *YTMusic) Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	results, err := y.searchOnce(ctx, query)
	if err != nil && errors.Is(err, errStaleCredentials) {
		y.invalidateCredentials()
		results, err = y.searchOnce(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

var errStaleCredentials = errors.New("stale innertube credentials")

func (y *YTMusic) searchOnce(ctx context.Context, query string) ([]domain.SearchResult, error) {
	apiKey, version, err := y.credentials(ctx)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "WEB_REMIX",
				"clientVersion": version,
				"hl":            "en",
				"gl":            "US",
			},
			"user": map[string]any{"lockedSafetyMode": false},
		},
		"query": query,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode search: %w", err)
	}
	url := y.baseURL + "?key=" + apiKey + "&prettyPrint=false"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://music.youtube.com")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := y.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden ||
		resp.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("%w: status %d", errStaleCredentials, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: search answered %d", ErrUnavailable, resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: read search response: %v", ErrUnavailable, err)
	}

	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("%w: search response is not JSON: %v", ErrUnavailable, err)
	}
	items := collectSongItems(doc)
	out := make([]domain.SearchResult, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if item.videoID == "" || seen[item.videoID] {
			continue
		}
		seen[item.videoID] = true
		out = append(out, domain.SearchResult{
			Track: domain.Track{
				Provider:    domain.ProviderYTMusic,
				ProviderID:  item.videoID,
				Title:       item.title,
				Artist:      item.artist,
				Album:       item.album,
				DurationMs:  item.durationMs,
				ArtworkURL:  item.artwork,
				ExternalURL: "https://music.youtube.com/watch?v=" + item.videoID,
			},
			ProviderName: "YouTube Music",
			Playable:     false,
			Note:         "Catalogo solamente: abre la pista en YouTube Music para escucharla",
		})
	}
	return out, nil
}

type songItem struct {
	videoID    string
	title      string
	artist     string
	album      string
	artwork    string
	durationMs int64
}

// collectSongItems walks the response tree rather than binding exact structs:
// InnerTube reshapes its renderers regularly, and a walk keeps working across
// those changes.
func collectSongItems(node any) []songItem {
	var out []songItem
	var walk func(n any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if pid, ok := v["playlistItemData"].(map[string]any); ok {
				if videoID, ok := pid["videoId"].(string); ok && videoID != "" {
					if item, ok := parseSongNode(v, videoID); ok {
						out = append(out, item)
					}
				}
			}
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(node)
	return out
}

func parseSongNode(node map[string]any, videoID string) (songItem, bool) {
	columns, _ := node["flexColumns"].([]any)
	texts := make([]string, 0, len(columns))
	for _, col := range columns {
		texts = append(texts, runsText(col))
	}
	item := songItem{videoID: videoID}
	if len(texts) > 0 {
		item.title = strings.TrimSpace(texts[0])
	}
	if len(texts) > 1 {
		item.artist = cleanSubtitle(texts[1])
	}
	if len(texts) > 2 {
		item.album = cleanSubtitle(texts[2])
	}
	if item.title == "" && len(texts) > 1 {
		item.title = strings.TrimSpace(texts[1])
	}
	if item.title == "" {
		return songItem{}, false
	}
	item.artwork = largestThumbnail(node)
	item.durationMs = durationFromNode(node)
	return item, true
}

// runsText concatenates every run of text inside a renderer node.
func runsText(node any) string {
	var b strings.Builder
	var walk func(n any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if text, ok := v["text"].(string); ok {
				b.WriteString(text)
			}
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(node)
	return b.String()
}

// resultKinds are the labels YouTube puts in the second column of a search
// row. They describe the result, they are not the artist.
var resultKinds = map[string]bool{
	"song": true, "video": true, "episode": true, "single": true,
	"album": true, "playlist": true, "artist": true, "podcast": true,
	"profile": true,
}

func cleanSubtitle(s string) string {
	s = strings.TrimSpace(s)
	// Separators between artist, album and duration arrive as bullets.
	if idx := strings.Index(s, " • "); idx > 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	if resultKinds[lower] {
		return ""
	}
	// "267K plays" / "1.2M views" are counters, not an album.
	if strings.HasSuffix(lower, " plays") || strings.HasSuffix(lower, " views") {
		return ""
	}
	return s
}

func largestThumbnail(node any) string {
	var best string
	var bestWidth float64
	var walk func(n any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if thumbs, ok := v["thumbnails"].([]any); ok {
				for _, t := range thumbs {
					m, ok := t.(map[string]any)
					if !ok {
						continue
					}
					url, _ := m["url"].(string)
					width, _ := m["width"].(float64)
					if url != "" && width >= bestWidth {
						best, bestWidth = url, width
					}
				}
			}
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(node)
	return best
}

func durationFromNode(node any) int64 {
	var out int64
	var walk func(n any)
	walk = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if txt, ok := v["text"].(string); ok && out == 0 {
				if ms, ok := parseClock(txt); ok {
					out = ms
				}
			}
			for _, child := range v {
				walk(child)
			}
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(node)
	return out
}

// parseClock converts "3:54" or "1:02:33" into milliseconds.
func parseClock(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, ":") || len(s) > 8 {
		return 0, false
	}
	parts := strings.Split(s, ":")
	for _, p := range parts {
		if p == "" {
			return 0, false
		}
		if _, err := strconv.Atoi(p); err != nil {
			return 0, false
		}
	}
	var total int64
	for _, p := range parts {
		n, _ := strconv.Atoi(p)
		total = total*60 + int64(n)
	}
	return total * 1000, true
}
