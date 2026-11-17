package demo

import (
	"context"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/media"
	"github.com/LuiisDev21/stefani/internal/storage"
)

func TestBuildProducesMonotonicWordTimings(t *testing.T) {
	for _, song := range Songs {
		built := build(song)
		if got, want := len(built.lyrics.Lines), len(song.Lines); got != want {
			t.Fatalf("%s: %d lines, want %d", song.Slug, got, want)
		}
		if built.lyrics.Sync != domain.SyncWord {
			t.Fatalf("%s: sync level = %s", song.Slug, built.lyrics.Sync)
		}
		var lastEnd int64
		for _, line := range built.lyrics.Lines {
			if line.StartMs < lastEnd {
				t.Fatalf("%s: line %q starts before the previous one ends", song.Slug, line.Text)
			}
			if len(line.Words) == 0 {
				t.Fatalf("%s: line %q has no word timings", song.Slug, line.Text)
			}
			var wordEnd int64
			for _, w := range line.Words {
				if w.StartMs < wordEnd {
					t.Fatalf("%s: word %q overlaps the previous word", song.Slug, w.Text)
				}
				if w.EndMs <= w.StartMs {
					t.Fatalf("%s: word %q has a non-positive duration", song.Slug, w.Text)
				}
				wordEnd = w.EndMs
			}
			if line.EndMs < wordEnd {
				t.Fatalf("%s: line %q ends before its last word", song.Slug, line.Text)
			}
			lastEnd = line.EndMs
		}
		if built.track.DurationMs <= lastEnd {
			t.Fatalf("%s: audio (%d ms) is shorter than the lyrics (%d ms)", song.Slug, built.track.DurationMs, lastEnd)
		}
	}
}

func TestEnsurePersistsAudioLyricsAndAnnotations(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(filepath.Join(t.TempDir(), "demo.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	lib := New(filepath.Join(t.TempDir(), "audio"))
	if err := lib.Ensure(ctx, db); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	for _, song := range Songs {
		track, err := db.Track(ctx, "demo:"+song.Slug)
		if err != nil {
			t.Fatalf("load demo track %s: %v", song.Slug, err)
		}
		if track.AudioPath == "" {
			t.Fatalf("%s: no audio path", song.Slug)
		}
		lyrics, err := db.Lyrics(ctx, track.ID)
		if err != nil {
			t.Fatalf("load lyrics for %s: %v", song.Slug, err)
		}
		if !lyrics.HasWordTimings() {
			t.Fatalf("%s: lyrics lost word timings", song.Slug)
		}
		annotations, err := db.AnnotationsByTrack(ctx, track.ID)
		if err != nil {
			t.Fatalf("load annotations for %s: %v", song.Slug, err)
		}
		if len(annotations) != len(song.Lines) {
			t.Fatalf("%s: %d annotations, want %d", song.Slug, len(annotations), len(song.Lines))
		}
	}
}

func TestEnsureIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(filepath.Join(t.TempDir(), "demo.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	lib := New(filepath.Join(t.TempDir(), "audio"))
	for i := 0; i < 3; i++ {
		if err := lib.Ensure(ctx, db); err != nil {
			t.Fatalf("ensure run %d: %v", i, err)
		}
	}
	tracks, err := db.AllTracks(ctx)
	if err != nil {
		t.Fatalf("all tracks: %v", err)
	}
	if len(tracks) != len(Songs) {
		t.Fatalf("%d tracks after repeated runs, want %d", len(tracks), len(Songs))
	}
	annotations, err := db.AnnotationsByTrack(ctx, "demo:"+Songs[0].Slug)
	if err != nil {
		t.Fatalf("annotations: %v", err)
	}
	if len(annotations) != len(Songs[0].Lines) {
		t.Fatalf("%d annotations after repeated runs, want %d", len(annotations), len(Songs[0].Lines))
	}
}

func TestRenderedWAVHasValidHeader(t *testing.T) {
	built := build(Songs[0])
	// Render a shortened version to keep the test fast.
	built.spec.Duration = 2
	data := media.RenderWAV(built.spec)
	if len(data) < 44 {
		t.Fatalf("wav too short: %d bytes", len(data))
	}
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		t.Fatalf("not a RIFF/WAVE file: %q %q", data[0:4], data[8:12])
	}
	channels := binary.LittleEndian.Uint16(data[22:24])
	if channels != 1 {
		t.Fatalf("channels = %d, want mono", channels)
	}
	rate := binary.LittleEndian.Uint32(data[24:28])
	if rate != SampleRate {
		t.Fatalf("sample rate = %d, want %d", rate, SampleRate)
	}
}
