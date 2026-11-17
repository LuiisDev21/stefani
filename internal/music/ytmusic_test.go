package music

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LuiisDev21/stefani/internal/domain"
)

func TestParseClock(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"3:54", 234000, true},
		{"1:02:33", 3753000, true},
		{"0:07", 7000, true},
		{"abc", 0, false},
		{"3:54 (Official Video)", 0, false},
	}
	for _, c := range cases {
		got, ok := parseClock(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Fatalf("parseClock(%q) = %d,%v want %d,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

// fakeInnerTube serves the two endpoints the client needs: the web page with
// the client config, and the search response.
func fakeInnerTube(t *testing.T) (*YTMusic, *int) {
	t.Helper()
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/" {
			w.Write([]byte(`<script>ytcfg.set({"INNERTUBE_API_KEY":"test-key","INNERTUBE_CLIENT_VERSION":"1.20250101.00.00"});</script>`))
			return
		}
		if r.URL.Path == "/youtubei/v1/search" {
			if r.URL.Query().Get("key") != "test-key" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			resp := map[string]any{
				"contents": map[string]any{
					"tabbedSearchResultsRenderer": map[string]any{
						"tabs": []any{map[string]any{
							"tabRenderer": map[string]any{
								"content": map[string]any{
									"sectionListRenderer": map[string]any{
										"contents": []any{map[string]any{
											"musicShelfRenderer": map[string]any{
												"contents": []any{
													songNode("vid1", "Take It Slowly", "Stefani Studio", "3:54", 640),
													songNode("vid2", "Paper Boats", "Stefani Studio", "2:31", 320),
												},
											},
										}},
									},
								},
							},
						}},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := NewYTMusic(0).WithBaseURL(srv.URL+"/youtubei/v1/search", srv.URL)
	return client, &calls
}

func songNode(id, title, artist, clock string, thumbWidth int) map[string]any {
	return map[string]any{
		"musicResponsiveListItemRenderer": map[string]any{
			"playlistItemData": map[string]any{"videoId": id},
			"flexColumns": []any{
				column(title),
				column(artist + " • " + "Demo Sessions"),
			},
			"thumbnail": map[string]any{
				"musicThumbnailRenderer": map[string]any{
					"thumbnail": map[string]any{
						"thumbnails": []any{
							map[string]any{"url": "https://img.example/small.jpg", "width": 60.0},
							map[string]any{"url": "https://img.example/large.jpg", "width": float64(thumbWidth)},
						},
					},
				},
			},
			"fixedColumns": []any{
				column(clock),
			},
		},
	}
}

func column(text string) map[string]any {
	return map[string]any{
		"musicResponsiveListItemFlexColumnRenderer": map[string]any{
			"text": map[string]any{"runs": []any{map[string]any{"text": text}}},
		},
	}
}

func TestYTMusicSearchParsesMetadata(t *testing.T) {
	client, _ := fakeInnerTube(t)
	results, err := client.Search(context.Background(), "slow", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	first := results[0]
	if first.Track.Provider != domain.ProviderYTMusic || first.Track.ProviderID != "vid1" {
		t.Fatalf("unexpected identity: %+v", first.Track)
	}
	if first.Track.Title != "Take It Slowly" {
		t.Fatalf("title = %q", first.Track.Title)
	}
	if first.Track.Artist != "Stefani Studio" {
		t.Fatalf("artist = %q", first.Track.Artist)
	}
	if first.Track.DurationMs != 234000 {
		t.Fatalf("duration = %d, want 234000", first.Track.DurationMs)
	}
	if !strings.Contains(first.Track.ArtworkURL, "large") {
		t.Fatalf("artwork = %q, want the widest thumbnail", first.Track.ArtworkURL)
	}
	if first.Playable {
		t.Fatal("catalog entries must not claim they are playable")
	}
	if !strings.Contains(first.Track.ExternalURL, "vid1") {
		t.Fatalf("external url = %q", first.Track.ExternalURL)
	}
}

func TestYTMusicRefreshesCredentialsOnce(t *testing.T) {
	client, calls := fakeInnerTube(t)
	// Seed a stale key: the first search must fail, refresh and succeed.
	client.apiKey = "stale"
	client.clientVersion = "1.20250101.00.00"
	results, err := client.Search(context.Background(), "slow", 10)
	if err != nil {
		t.Fatalf("search with stale key: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results after refresh, want 2", len(results))
	}
	if *calls < 3 {
		t.Fatalf("expected a page fetch plus two search calls, got %d calls", *calls)
	}
}
