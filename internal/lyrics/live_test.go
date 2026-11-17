package lyrics

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/storage"
)

// TestLiveLookupForOnlineTrack checks the real LRCLIB service with the metadata
// shape YouTube Music returns (duration is often 0, artist may be a channel).
func TestLiveLookupForOnlineTrack(t *testing.T) {
	if os.Getenv("STEFANI_LIVE") != "1" {
		t.Skip("set STEFANI_LIVE=1 to query LRCLIB")
	}
	db, err := storage.Open(filepath.Join(t.TempDir(), "live.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	svc := NewService(db, NewLRCLIB(15*time.Second), noopLogger{})

	cases := []domain.Track{
		{ID: "ytmusic:dQw4w9WgXcQ", Provider: domain.ProviderYTMusic, ProviderID: "dQw4w9WgXcQ",
			Title: "Never Gonna Give You Up", Artist: "Rick Astley", DurationMs: 213000},
		{ID: "ytmusic:zero", Provider: domain.ProviderYTMusic, ProviderID: "zero",
			Title: "Never Gonna Give You Up", Artist: "Rick Astley", DurationMs: 0},
		{ID: "ytmusic:channel", Provider: domain.ProviderYTMusic, ProviderID: "channel",
			Title: "Bohemian Rhapsody", Artist: "Queen Official", DurationMs: 0},
	}
	for _, track := range cases {
		id, err := db.SaveTrack(context.Background(), track)
		if err != nil {
			t.Fatalf("save track: %v", err)
		}
		track.ID = id
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		doc, err := svc.ForTrack(ctx, track, true)
		cancel()
		if err != nil {
			t.Logf("%s -> error: %v", track.ID, err)
			continue
		}
		wordSync := 0
		for _, line := range doc.Lines {
			if len(line.Words) > 0 {
				wordSync++
			}
		}
		t.Logf("%s -> %d lines, sync=%s, words_on=%d, source=%s", track.ID, len(doc.Lines), doc.Sync, wordSync, doc.Source)
	}
}

type noopLogger struct{}

func (noopLogger) Info(msg string, args ...any) {}
func (noopLogger) Warn(msg string, args ...any) {}

func TestLiveLookupWithoutArtist(t *testing.T) {
	if os.Getenv("STEFANI_LIVE") != "1" {
		t.Skip("set STEFANI_LIVE=1 to hit the real LRCLIB API")
	}
	lib := NewLRCLIB(20 * time.Second)
	// The exact shape a YouTube Music row produces: decorated title, no artist.
	doc, err := lib.Lookup(context.Background(), domain.Track{
		Title: "Never Gonna Give You Up (Official Video)",
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if doc == nil || stringValue(doc.SyncedLyrics) == "" {
		t.Fatalf("no synced lyrics: %+v", doc)
	}
	t.Logf("%s — %s (%d lines)", doc.TrackName, doc.ArtistName, strings.Count(stringValue(doc.SyncedLyrics), "\n")+1)
}
