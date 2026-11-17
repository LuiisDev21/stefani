package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/LuiisDev21/stefani/internal/domain"
)

func TestSearchTracksHidesIncompleteOnlineRows(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "search.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()

	// A row cached by the old, unfiltered search: a podcast with no artist and
	// no duration. It must never surface as a song again.
	if _, err := db.SaveTrack(ctx, domain.Track{
		Provider: domain.ProviderYTMusic, ProviderID: "podcast1",
		Title: "Never Gonna Give You Up (podcast episode)",
	}); err != nil {
		t.Fatalf("save podcast: %v", err)
	}
	if _, err := db.SaveTrack(ctx, domain.Track{
		Provider: domain.ProviderYTMusic, ProviderID: "song1",
		Title: "Never Gonna Give You Up", Artist: "Rick Astley", DurationMs: 214000,
	}); err != nil {
		t.Fatalf("save song: %v", err)
	}

	got, err := db.SearchTracks(ctx, "never gonna", "", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 || got[0].ProviderID != "song1" {
		t.Fatalf("results = %+v, want only the complete song", got)
	}
}
