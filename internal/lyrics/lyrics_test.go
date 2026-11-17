package lyrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/storage"
)

func TestParsePlainTextIsUnsynced(t *testing.T) {
	parsed := Parse("First line\n\nSecond line\n")
	if parsed.Sync != domain.SyncUnsynced {
		t.Fatalf("sync = %s, want UNSYNCED", parsed.Sync)
	}
	if len(parsed.Lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(parsed.Lines))
	}
	if parsed.Lines[0].StartMs != 0 || parsed.Lines[0].EndMs != 0 {
		t.Fatalf("unsynced lines must not carry guessed timings: %+v", parsed.Lines[0])
	}
}

func TestParseLRCUsesLineTimings(t *testing.T) {
	src := `[ti:Slow Morning]
[ar:Stefani Studio]
[00:12.50]Coffee steams against the glass
[00:17.00]Minutes move but do not pass
`
	parsed := Parse(src)
	if parsed.Title != "Slow Morning" || parsed.Artist != "Stefani Studio" {
		t.Fatalf("metadata not read: %+v", parsed)
	}
	if parsed.Sync != domain.SyncLine {
		t.Fatalf("sync = %s, want LINE", parsed.Sync)
	}
	if len(parsed.Lines) != 2 {
		t.Fatalf("got %d lines", len(parsed.Lines))
	}
	if parsed.Lines[0].StartMs != 12500 {
		t.Fatalf("start = %d, want 12500", parsed.Lines[0].StartMs)
	}
	if parsed.Lines[0].EndMs != 17000 {
		t.Fatalf("end = %d, want 17000 (next line start)", parsed.Lines[0].EndMs)
	}
	if len(parsed.Lines[0].Words) != 0 {
		t.Fatal("line timings must not be turned into fake word timings")
	}
}

func TestParseEnhancedLRCBuildsWordTimings(t *testing.T) {
	src := `[00:10.00]<00:10.00>Take <00:10.45>it <00:10.90>slowly`
	parsed := Parse(src)
	if parsed.Sync != domain.SyncWord {
		t.Fatalf("sync = %s, want WORD", parsed.Sync)
	}
	if len(parsed.Lines) != 1 {
		t.Fatalf("got %d lines", len(parsed.Lines))
	}
	line := parsed.Lines[0]
	if line.Text != "Take it slowly" {
		t.Fatalf("text = %q", line.Text)
	}
	if len(line.Words) != 3 {
		t.Fatalf("got %d words", len(line.Words))
	}
	if line.Words[0].Text != "Take" || line.Words[0].StartMs != 10000 {
		t.Fatalf("first word: %+v", line.Words[0])
	}
	if line.Words[1].StartMs != 10450 {
		t.Fatalf("second word start = %d", line.Words[1].StartMs)
	}
	if line.Words[0].EndMs != 10450 {
		t.Fatalf("word end must equal the next word start: %+v", line.Words[0])
	}
}

func TestParseOffsetShiftsTimings(t *testing.T) {
	src := `[offset:-500]
[00:10.00]first
[00:12.00]second`
	parsed := Parse(src)
	if parsed.OffsetMs != -500 {
		t.Fatalf("offset = %d", parsed.OffsetMs)
	}
	if parsed.Lines[0].StartMs != 9500 {
		t.Fatalf("start = %d, want 9500", parsed.Lines[0].StartMs)
	}
}

func TestParseDuplicatedTimestampsExpandLines(t *testing.T) {
	src := `[00:10.00][01:10.00]same chorus`
	parsed := Parse(src)
	if len(parsed.Lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(parsed.Lines))
	}
	if parsed.Lines[0].StartMs != 10000 || parsed.Lines[1].StartMs != 70000 {
		t.Fatalf("timestamps = %d, %d", parsed.Lines[0].StartMs, parsed.Lines[1].StartMs)
	}
}

func TestParseSortsOutOfOrderLines(t *testing.T) {
	src := "[00:20.00]second\n[00:10.00]first"
	parsed := Parse(src)
	if parsed.Lines[0].Text != "first" || parsed.Lines[1].Text != "second" {
		t.Fatalf("lines not sorted: %+v", parsed.Lines)
	}
}

func TestValidateRejectsOverlap(t *testing.T) {
	doc := domain.Lyrics{Lines: []domain.Line{
		{StartMs: 0, EndMs: 5000},
		{StartMs: 4000, EndMs: 8000},
	}}
	if err := Validate(doc); err == nil {
		t.Fatal("expected an overlap error")
	}
}

func fakeLRCLIB(t *testing.T, searchResults []map[string]any) (*LRCLIB, *int) {
	t.Helper()
	hits := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{}`))
	})
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		hits++
		json.NewEncoder(w).Encode(searchResults)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return NewLRCLIB(0).WithBaseURL(srv.URL), &hits
}

func TestLookupFallsBackToSearchAndMatchesDuration(t *testing.T) {
	lib, hits := fakeLRCLIB(t, []map[string]any{
		{
			"id": 1.0, "trackName": "Slow Morning", "artistName": "Stefani Studio",
			"duration": 210.0, "syncedLyrics": "[00:01.00]wrong version",
		},
		{
			"id": 2.0, "trackName": "Slow Morning", "artistName": "Stefani Studio",
			"duration": 36.0, "syncedLyrics": "[00:01.00]right version",
		},
	})
	got, err := lib.Lookup(context.Background(), domain.Track{
		Title: "Slow Morning", Artist: "Stefani Studio", DurationMs: 36500,
	})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if got.ID != 2 {
		t.Fatalf("picked record %d, want the duration match (2)", got.ID)
	}
	if *hits != 2 {
		t.Fatalf("expected an exact attempt plus a search, got %d requests", *hits)
	}
}

func TestServiceCachesRemoteLyrics(t *testing.T) {
	lib, hits := fakeLRCLIB(t, []map[string]any{
		{
			"id": 9.0, "trackName": "Paper Boats", "artistName": "Stefani Studio",
			"duration": 40.0, "syncedLyrics": "[00:01.00]<00:01.00>We <00:01.50>folded <00:02.00>words",
		},
	})
	db, err := storage.Open(filepath.Join(t.TempDir(), "lyrics.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	track := domain.Track{ID: "local:t1", Provider: domain.ProviderLocal, ProviderID: "t1", Title: "Paper Boats", Artist: "Stefani Studio", DurationMs: 40000}
	if _, err := db.SaveTrack(context.Background(), track); err != nil {
		t.Fatalf("save track: %v", err)
	}
	service := NewService(db, lib, testLogger{t})

	first, err := service.ForTrack(context.Background(), track, false)
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	if first.Sync != domain.SyncWord {
		t.Fatalf("sync = %s, want WORD", first.Sync)
	}
	requestsAfterFirst := *hits
	second, err := service.ForTrack(context.Background(), track, false)
	if err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	if *hits != requestsAfterFirst {
		t.Fatalf("cached fetch hit the network again: %d -> %d", requestsAfterFirst, *hits)
	}
	if len(second.Lines) != len(first.Lines) {
		t.Fatalf("cached document differs: %d vs %d lines", len(second.Lines), len(first.Lines))
	}
}

func TestServiceReadsSidecarFile(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "song.mp3")
	if err := writeFile(audio, []byte("not really audio")); err != nil {
		t.Fatal(err)
	}
	lrc := strings.TrimSuffix(audio, ".mp3") + ".lrc"
	if err := writeFile(lrc, []byte("[00:01.00]hello\n[00:03.00]world")); err != nil {
		t.Fatal(err)
	}
	db, err := storage.Open(filepath.Join(dir, "sidecar.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	track := domain.Track{ID: "local:s1", Provider: domain.ProviderLocal, ProviderID: "s1", Title: "Song", AudioPath: audio}
	if _, err := db.SaveTrack(context.Background(), track); err != nil {
		t.Fatalf("save: %v", err)
	}
	service := NewService(db, nil, testLogger{t})
	doc, err := service.ForTrack(context.Background(), track, false)
	if err != nil {
		t.Fatalf("sidecar lyrics: %v", err)
	}
	if doc.Provider != "file" || len(doc.Lines) != 2 {
		t.Fatalf("unexpected document: %+v", doc)
	}
}

type testLogger struct{ t *testing.T }

func (l testLogger) Info(msg string, args ...any) { l.t.Logf(msg, args...) }
func (l testLogger) Warn(msg string, args ...any) { l.t.Logf("WARN: "+msg, args...) }

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
