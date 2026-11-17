package music

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// browseURL is the InnerTube endpoint that serves playlist pages.
const browseURL = "https://music.youtube.com/youtubei/v1/browse"

// WithBrowseURL overrides the browse endpoint, for tests.
func (y *YTMusic) WithBrowseURL(u string) *YTMusic {
	y.browse = u
	return y
}

// BrowsePlaylist reads the public metadata of a playlist (no account, no
// cookies). Private and region-locked playlists simply return nothing.
func (y *YTMusic) BrowsePlaylist(ctx context.Context, playlistID string, limit int) ([]domain.Track, error) {
	playlistID = strings.TrimSpace(playlistID)
	if playlistID == "" {
		return nil, fmt.Errorf("empty playlist id")
	}
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	tracks, err := y.browseOnce(ctx, playlistID)
	if err != nil && err == errStaleCredentials {
		y.invalidateCredentials()
		tracks, err = y.browseOnce(ctx, playlistID)
	}
	if err != nil {
		return nil, err
	}
	if len(tracks) > limit {
		tracks = tracks[:limit]
	}
	return tracks, nil
}

func (y *YTMusic) browseOnce(ctx context.Context, playlistID string) ([]domain.Track, error) {
	apiKey, version, err := y.credentials(ctx)
	if err != nil {
		return nil, err
	}
	browse := y.browse
	if browse == "" {
		browse = browseURL
	}
	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "WEB_REMIX",
				"clientVersion": version,
				"hl":            "en",
				"gl":            "US",
			},
		},
		"browseId": "VL" + playlistID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode browse: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		browse+"?key="+apiKey+"&prettyPrint=false", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build browse request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://music.youtube.com")

	resp, err := y.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden ||
		resp.StatusCode == http.StatusBadRequest {
		return nil, errStaleCredentials
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: browse answered %d", ErrUnavailable, resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: read browse response: %v", ErrUnavailable, err)
	}
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("%w: browse response is not JSON: %v", ErrUnavailable, err)
	}
	items := collectSongItems(doc)
	out := make([]domain.Track, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if item.videoID == "" || seen[item.videoID] {
			continue
		}
		seen[item.videoID] = true
		out = append(out, domain.Track{
			Provider:    domain.ProviderYTMusic,
			ProviderID:  item.videoID,
			Title:       item.title,
			Artist:      item.artist,
			Album:       item.album,
			DurationMs:  item.durationMs,
			ArtworkURL:  item.artwork,
			ExternalURL: "https://music.youtube.com/watch?v=" + item.videoID,
		})
	}
	return out, nil
}

// ImportPlaylist exposes playlist import through the service.
func (s *Service) ImportPlaylist(ctx context.Context, playlistID string) ([]domain.Track, error) {
	if s.yt == nil {
		return nil, fmt.Errorf("el catalogo de YouTube Music no esta configurado")
	}
	return s.yt.BrowsePlaylist(ctx, playlistID, 200)
}
