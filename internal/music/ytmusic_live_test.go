package music

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestYTMusicLiveSearch hits the real endpoint. It is skipped unless the
// developer opts in with STEFANI_LIVE=1, so CI stays deterministic.
func TestYTMusicLiveSearch(t *testing.T) {
	if os.Getenv("STEFANI_LIVE") == "" {
		t.Skip("set STEFANI_LIVE=1 to query the real YouTube Music endpoint")
	}
	client := NewYTMusic(20 * time.Second)
	results, err := client.Search(context.Background(), "bohemian rhapsody", 5)
	if err != nil {
		t.Fatalf("live search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("live search returned no results; the endpoint contract changed")
	}
	first := results[0]
	t.Logf("first hit: %q by %q (%d ms) id=%s artwork=%s",
		first.Track.Title, first.Track.Artist, first.Track.DurationMs, first.Track.ProviderID, first.Track.ArtworkURL)
	if first.Track.ProviderID == "" || first.Track.Title == "" {
		t.Fatalf("incomplete metadata: %+v", first.Track)
	}
}
