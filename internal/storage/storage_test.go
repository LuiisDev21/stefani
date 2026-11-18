package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/LuiisDev21/stefani/internal/domain"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedTrack(t *testing.T, db *DB, id string) {
	t.Helper()
	_, err := db.SaveTrack(context.Background(), domain.Track{
		ID:         id,
		Provider:   domain.ProviderLocal,
		ProviderID: id,
		Title:      "Track " + id,
		Artist:     "Tester",
		DurationMs: 1000,
	})
	if err != nil {
		t.Fatalf("seed track %s: %v", id, err)
	}
}

func TestMigrationsApplyOnce(t *testing.T) {
	db := newTestDB(t)
	version, err := db.SchemaVersion(context.Background())
	if err != nil {
		t.Fatalf("schema version: %v", err)
	}
	if version != 2 {
		t.Fatalf("schema version = %d, want 2", version)
	}
	// Re-opening the same file must not re-apply migrations.
	path := db.Path()
	db.Close()
	again, err := Open(path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer again.Close()
	if v, _ := again.SchemaVersion(context.Background()); v != 2 {
		t.Fatalf("schema version after reopen = %d, want 2", v)
	}
}

func TestForeignKeysAreEnforced(t *testing.T) {
	db := newTestDB(t)
	_, err := db.sql.Exec(`INSERT INTO library_tracks (track_id, is_favorite, saved_at) VALUES ('missing', 0, 1)`)
	if err == nil {
		t.Fatal("expected a foreign key violation for an unknown track")
	}
}

func TestTrackUpsertKeepsSingleRow(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	id, err := db.SaveTrack(ctx, domain.Track{
		Provider: domain.ProviderLocal, ProviderID: "abc", Title: "First", Artist: "A",
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	_, err = db.SaveTrack(ctx, domain.Track{
		Provider: domain.ProviderLocal, ProviderID: "abc", Title: "Second", Artist: "B",
	})
	if err != nil {
		t.Fatalf("resave: %v", err)
	}
	got, err := db.Track(ctx, id)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Title != "Second" || got.Artist != "B" {
		t.Fatalf("track not updated: %+v", got)
	}
	all, err := db.AllTracks(ctx)
	if err != nil {
		t.Fatalf("all tracks: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("got %d tracks, want 1", len(all))
	}
}

func TestPlaylistOrderingAndMove(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	for _, id := range []string{"t1", "t2", "t3"} {
		seedTrack(t, db, id)
	}
	pl, err := db.CreatePlaylist(ctx, domain.Playlist{Name: "Practice"})
	if err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	for _, id := range []string{"t1", "t2", "t3"} {
		if err := db.AddToPlaylist(ctx, pl.ID, id); err != nil {
			t.Fatalf("add %s: %v", id, err)
		}
	}

	items, err := db.PlaylistItems(ctx, pl.ID)
	if err != nil {
		t.Fatalf("items: %v", err)
	}
	wantOrder(t, items, []string{"t1", "t2", "t3"})

	if err := db.MovePlaylistItem(ctx, pl.ID, 0, 2); err != nil {
		t.Fatalf("move: %v", err)
	}
	items, _ = db.PlaylistItems(ctx, pl.ID)
	wantOrder(t, items, []string{"t2", "t3", "t1"})

	if err := db.RemovePlaylistItem(ctx, pl.ID, 1); err != nil {
		t.Fatalf("remove: %v", err)
	}
	items, _ = db.PlaylistItems(ctx, pl.ID)
	wantOrder(t, items, []string{"t2", "t1"})
	for i, item := range items {
		if item.Position != i {
			t.Fatalf("position gap after remove: %+v", items)
		}
	}
}

func wantOrder(t *testing.T, items []domain.PlaylistItem, want []string) {
	t.Helper()
	if len(items) != len(want) {
		t.Fatalf("got %d items, want %d", len(items), len(want))
	}
	for i, id := range want {
		if items[i].Track.ID != id {
			t.Fatalf("position %d = %s, want %s", i, items[i].Track.ID, id)
		}
	}
}

func TestSaveLyricsKeepsAnnotationsForUnchangedLines(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	seedTrack(t, db, "t1")

	doc := domain.Lyrics{
		TrackID:  "t1",
		Provider: "test",
		Sync:     domain.SyncWord,
		Lines: []domain.Line{
			{ID: "t1-l0", Text: "hello world", StartMs: 0, EndMs: 1000, Words: []domain.Word{
				{Text: "hello", StartMs: 0, EndMs: 500}, {Text: "world", StartMs: 500, EndMs: 1000},
			}},
			{ID: "t1-l1", Text: "second line", StartMs: 1000, EndMs: 2000},
		},
	}
	if err := db.SaveLyrics(ctx, doc); err != nil {
		t.Fatalf("save lyrics: %v", err)
	}
	if err := db.SaveAnnotation(ctx, domain.Annotation{
		LineID: "t1-l0", Accent: "en-US", IPA: "/həˈloʊ wɜːrld/", SpanishApprox: "jeló uerld",
	}); err != nil {
		t.Fatalf("save annotation: %v", err)
	}

	// Rebuild with one changed line: the untouched line keeps its annotation.
	doc.Lines[1].Text = "second verse"
	if err := db.SaveLyrics(ctx, doc); err != nil {
		t.Fatalf("resave lyrics: %v", err)
	}
	if _, err := db.AnnotationForLine(ctx, "t1-l0", "en-US"); err != nil {
		t.Fatalf("annotation for unchanged line was lost: %v", err)
	}
	loaded, err := db.Lyrics(ctx, "t1")
	if err != nil {
		t.Fatalf("load lyrics: %v", err)
	}
	if len(loaded.Lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(loaded.Lines))
	}
	if len(loaded.Lines[0].Words) != 2 {
		t.Fatalf("word timings lost: %+v", loaded.Lines[0])
	}
	if loaded.Lines[1].Text != "second verse" {
		t.Fatalf("line text not updated: %q", loaded.Lines[1].Text)
	}
}

func TestDeletePlaylistCascadesItems(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	seedTrack(t, db, "t1")
	pl, _ := db.CreatePlaylist(ctx, domain.Playlist{Name: "Temp"})
	if err := db.AddToPlaylist(ctx, pl.ID, "t1"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := db.DeletePlaylist(ctx, pl.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	var count int
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM playlist_items`).Scan(&count); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if count != 0 {
		t.Fatalf("playlist items survived cascade: %d", count)
	}
}
