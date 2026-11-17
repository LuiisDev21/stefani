package lyrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
		if r.URL.Query().Get("artist_name") == "" {
			// LRCLIB's actual behaviour: /get without artist_name is a 400.
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Failed to deserialize query string: missing field artist_name"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{}`))
	})
	// /get/{id}: the v2 flow fetches only the chosen record.
	mux.HandleFunc("/get/", func(w http.ResponseWriter, r *http.Request) {
		hits++
		id := strings.TrimPrefix(r.URL.Path, "/get/")
		for _, item := range searchResults {
			if fmt.Sprint(item["id"]) == id {
				json.NewEncoder(w).Encode(item)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Failed to find specified track","name":"TrackNotFound","statusCode":404}`))
	})
	// /v2/search: summaries only, exactly as the beta endpoint documents.
	mux.HandleFunc("/v2/search", func(w http.ResponseWriter, r *http.Request) {
		hits++
		out := make([]map[string]any, 0, len(searchResults))
		for _, item := range searchResults {
			if r.URL.Query().Get("track_name") != "" {
				title, _ := item["trackName"].(string)
				if !strings.Contains(strings.ToLower(title), strings.ToLower(r.URL.Query().Get("track_name"))) {
					continue
				}
			}
			summary := map[string]any{
				"id":              item["id"],
				"trackName":       item["trackName"],
				"artistName":      item["artistName"],
				"duration":        item["duration"],
				"instrumental":    item["instrumental"],
				"hasWordSync":     item["hasWordSync"],
				"hasSyncedLyrics": item["syncedLyrics"] != nil,
				"hasPlainLyrics":  item["plainLyrics"] != nil,
				"hasLyricsfile":   true,
			}
			out = append(out, summary)
		}
		json.NewEncoder(w).Encode(out)
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
	// One exact attempt, one v2 search and one fetch of the chosen record.
	if *hits != 3 {
		t.Fatalf("expected exact attempt + v2 search + record fetch, got %d requests", *hits)
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

func TestLookupWithoutCandidatesIsNoLyrics(t *testing.T) {
	// The search works but nothing matches: the caller must get ErrNoLyrics,
	// never a nil document it would dereference.
	lib, _ := fakeLRCLIB(t, []map[string]any{})
	got, err := lib.Lookup(context.Background(), domain.Track{
		Title: "Never Gonna Give You Up", Artist: "Rick Astley", DurationMs: 0,
	})
	if !errors.Is(err, ErrNoLyrics) {
		t.Fatalf("error = %v, want ErrNoLyrics", err)
	}
	if got != nil {
		t.Fatalf("document = %+v, want nil", got)
	}
}

func TestLookupIgnoresCandidatesWithoutLyrics(t *testing.T) {
	lib, _ := fakeLRCLIB(t, []map[string]any{
		{"id": 1.0, "trackName": "Never Gonna Give You Up", "artistName": "Rick Astley", "duration": 213.0},
		{"id": 2.0, "trackName": "Never Gonna Give You Up", "artistName": "Rick Astley", "duration": 213.0, "instrumental": true},
	})
	if _, err := lib.Lookup(context.Background(), domain.Track{
		Title: "Never Gonna Give You Up", Artist: "Rick Astley",
	}); !errors.Is(err, ErrNoLyrics) {
		t.Fatalf("error = %v, want ErrNoLyrics", err)
	}
}

func TestLookupWithoutArtistDoesNotFail(t *testing.T) {
	// A YouTube row can arrive with no artist at all. /api/get would answer 400,
	// and the lyric must still be found through the search fallback.
	lib, _ := fakeLRCLIB(t, []map[string]any{
		{
			"id": 1.0, "trackName": "Never Gonna Give You Up", "artistName": "Rick Astley",
			"duration": 213.0, "syncedLyrics": "[00:01.00]We're no strangers to love",
		},
	})
	got, err := lib.Lookup(context.Background(), domain.Track{Title: "Never Gonna Give You Up"})
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if got == nil || got.ArtistName != "Rick Astley" {
		t.Fatalf("document = %+v, want the Rick Astley record", got)
	}
}

func TestSearchBadRequestIsNoLyrics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)
	lib := NewLRCLIB(0).WithBaseURL(srv.URL)
	if _, err := lib.Lookup(context.Background(), domain.Track{Title: "x"}); !errors.Is(err, ErrNoLyrics) {
		t.Fatalf("error = %v, want ErrNoLyrics: a rejected query is not an outage", err)
	}
}

func TestCleanTitleRemovesVideoDecoration(t *testing.T) {
	cases := map[string]string{
		"Never Gonna Give You Up (Official Video)": "Never Gonna Give You Up",
		"Song [Lyrics]":          "Song",
		"Song (Remastered 2011)": "Song",
		"Song (Live at Wembley)": "Song",
		"Live and Let Die":       "Live and Let Die",
		"Song (feat. Someone)":   "Song (feat. Someone)",
	}
	for raw, want := range cases {
		if got := cleanTitle(raw); got != want {
			t.Fatalf("cleanTitle(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestCleanArtistRemovesTopicSuffix(t *testing.T) {
	cases := map[string]string{
		"Rick Astley - Topic": "Rick Astley",
		"  Queen  ":           "Queen",
		"":                    "",
	}
	for raw, want := range cases {
		if got := cleanArtist(raw); got != want {
			t.Fatalf("cleanArtist(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestLookupUsesV2SearchAndFetchesOnlyTheChosenRecord(t *testing.T) {
	var requested []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = append(requested, r.URL.Path)
		switch {
		case r.URL.Path == "/v2/search":
			w.Write([]byte(`[{"id":7,"trackName":"Slow Morning","artistName":"Stefani Studio","duration":210,"hasSyncedLyrics":true}]`))
		case r.URL.Path == "/get/7":
			w.Write([]byte(`{"id":7,"trackName":"Slow Morning","artistName":"Stefani Studio","duration":210,"syncedLyrics":"[00:01.00]one line"}`))
		case r.URL.Path == "/get":
			w.WriteHeader(http.StatusBadRequest)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	lib := NewLRCLIB(5 * time.Second).WithBaseURL(srv.URL)
	doc, err := lib.Lookup(context.Background(), domain.Track{Title: "Slow Morning (Official Video)", DurationMs: 210_000})
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if doc == nil || doc.ID != 7 {
		t.Fatalf("document = %+v, want record 7", doc)
	}
	if len(requested) < 2 || requested[0] != "/v2/search" || requested[1] != "/get/7" {
		t.Fatalf("request path = %v, want v2 search then the chosen record", requested)
	}
}

func TestRateLimitHonorsRetryAfter(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	lib := NewLRCLIB(5 * time.Second).WithBaseURL(srv.URL)
	var slept []time.Duration
	lib.sleep = func(d time.Duration) { slept = append(slept, d) }

	if _, err := lib.SearchSummaries(context.Background(), "x", "", ""); err != nil {
		t.Fatalf("search: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2 after the Retry-After retry", attempts)
	}
	if len(slept) != 1 || slept[0] < time.Second {
		t.Fatalf("waited %v, want at least the Retry-After second", slept)
	}
}

func TestLongRetryAfterFailsWithoutBlocking(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	lib := NewLRCLIB(5 * time.Second).WithBaseURL(srv.URL)
	lib.sleep = func(time.Duration) { t.Fatal("must not block for a two-minute Retry-After") }
	_, err := lib.SearchSummaries(context.Background(), "x", "", "")
	if err == nil || !strings.Contains(err.Error(), "retry in") {
		t.Fatalf("error = %v, want a readable rate-limit message", err)
	}
}

func TestParseRetryAfterAcceptsSecondsAndDates(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if got := parseRetryAfter("3", now); got != 3*time.Second {
		t.Fatalf("seconds = %v", got)
	}
	if got := parseRetryAfter(now.Add(5*time.Second).Format(http.TimeFormat), now); got != 5*time.Second {
		t.Fatalf("date = %v", got)
	}
	if got := parseRetryAfter("", now); got != 0 {
		t.Fatalf("empty = %v", got)
	}
}
