package demo

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/media"
	"github.com/LuiisDev21/stefani/internal/storage"
)

// SampleRate of the rendered demo audio.
const SampleRate = 44100

// AnnotationVersion stamps the precomputed study data.
const AnnotationVersion = "demo-v1"

// Library materializes the bundled demo songs: audio files on disk, tracks,
// word-timed lyrics and precomputed pronunciation data. It is idempotent, so it
// can run on every start.
type Library struct {
	dir string
}

// New points the library at the folder where demo audio is rendered.
func New(dir string) *Library { return &Library{dir: dir} }

// Dir returns the demo audio folder.
func (l *Library) Dir() string { return l.dir }

type builtSong struct {
	track   domain.Track
	lyrics  domain.Lyrics
	annot   []domain.Annotation
	spec    media.Spec
	durance float64
}

// Ensure renders and stores every demo song. Missing audio is rendered again;
// existing files are reused so start-up stays fast.
func (l *Library) Ensure(ctx context.Context, db *storage.DB) error {
	if err := os.MkdirAll(l.dir, 0o755); err != nil {
		return fmt.Errorf("create demo directory: %w", err)
	}
	for _, song := range Songs {
		built := build(song)
		path := filepath.Join(l.dir, song.Slug+".wav")
		if info, err := os.Stat(path); err != nil || info.Size() == 0 {
			if err := writeFileAtomic(path, media.RenderWAV(built.spec)); err != nil {
				return fmt.Errorf("render demo song %q: %w", song.Slug, err)
			}
		}
		built.track.AudioPath = path
		if _, err := db.SaveTrack(ctx, built.track); err != nil {
			return err
		}
		if err := db.SaveLyrics(ctx, built.lyrics); err != nil {
			return err
		}
		for _, a := range built.annot {
			if err := db.SaveAnnotation(ctx, a); err != nil {
				return err
			}
		}
	}
	return nil
}

// Tracks returns the demo tracks without touching the database, for projections
// and tests.
func Tracks() []domain.Track {
	out := make([]domain.Track, 0, len(Songs))
	for _, s := range Songs {
		built := build(s)
		out = append(out, built.track)
	}
	return out
}

func build(song Song) builtSong {
	spb := 60 / song.BPM
	bar := 4 * spb
	introBars := 2.0
	start := introBars * bar

	lines := make([]domain.Line, 0, len(song.Lines))
	type noteAt struct {
		startSec float64
		durSec   float64
		freq     float64
	}
	var melody []noteAt

	t := start
	for i, line := range song.Lines {
		lineStart := t
		words := make([]domain.Word, 0, len(line.Words))
		for j, word := range line.Words {
			dur := word.Beats * spb
			wordEnd := t + dur*0.94
			words = append(words, domain.Word{
				Text:      word.Text,
				StartMs:   int64(math.Round(t * 1000)),
				EndMs:     int64(math.Round(wordEnd * 1000)),
				TimingSrc: "generated",
			})
			degree := song.Mode[(i*3+j)%len(song.Mode)]
			octave := 0
			if j%4 == 3 {
				octave = 12
			}
			melody = append(melody, noteAt{
				startSec: t,
				durSec:   dur * 0.88,
				freq:     midiFreq(song.Root + degree + octave),
			})
			t += dur
		}
		// Round up to the next bar so each verse starts on a downbeat, then
		// leave half a bar of air before the next line.
		t = math.Ceil((t+bar*0.5)/bar) * bar
		lines = append(lines, domain.Line{
			ID:      fmt.Sprintf("demo:%s-l%03d", song.Slug, i),
			Index:   i,
			Text:    line.Text,
			StartMs: int64(math.Round(lineStart * 1000)),
			EndMs:   int64(math.Round((t - bar*0.5) * 1000)),
			Words:   words,
		})
	}

	duration := t + bar*2
	spec := media.Spec{SampleRate: SampleRate, Duration: duration}
	if spec.Duration < 8 {
		spec.Duration = 8
	}

	// Pads: one chord per bar, following the progression.
	for b := 0; ; b++ {
		barStart := float64(b) * bar
		if barStart >= duration {
			break
		}
		degree := song.Progression[b%len(song.Progression)]
		for _, voice := range []int{0, 2, 4} {
			d := song.Mode[(degree+voice)%len(song.Mode)]
			spec.Notes = append(spec.Notes, media.Note{
				StartSec: barStart,
				DurSec:   bar * 0.98,
				Freq:     midiFreq(song.Root - 12 + d),
				Amp:      0.10,
				Timbre:   media.TimbrePad,
			})
		}
		// Soft pulse on beats one and three.
		spec.Notes = append(spec.Notes,
			media.Note{StartSec: barStart, DurSec: 0.28, Freq: 60, Amp: 0.30, Timbre: media.TimbreKick},
			media.Note{StartSec: barStart + 2*spb, DurSec: 0.28, Freq: 60, Amp: 0.20, Timbre: media.TimbreKick},
		)
		for beat := 0; beat < 4; beat++ {
			spec.Notes = append(spec.Notes, media.Note{
				StartSec: barStart + float64(beat)*spb + spb*0.5,
				DurSec:   0.05,
				Amp:      0.035,
				Timbre:   media.TimbreHat,
			})
		}
	}
	for _, m := range melody {
		spec.Notes = append(spec.Notes, media.Note{
			StartSec: m.startSec,
			DurSec:   m.durSec,
			Freq:     m.freq,
			Amp:      0.22,
			Timbre:   media.TimbreMelody,
		})
	}

	annot := make([]domain.Annotation, 0, len(song.Lines))
	for i, line := range song.Lines {
		if line.IPA == "" {
			continue
		}
		notes := make([]domain.Note, 0, len(line.Notes))
		for _, n := range line.Notes {
			notes = append(notes, domain.Note{Type: "pronunciation", Text: n})
		}
		annot = append(annot, domain.Annotation{
			LineID:        fmt.Sprintf("demo:%s-l%03d", song.Slug, i),
			Accent:        "en-US",
			IPA:           line.IPA,
			SpanishApprox: line.Approx,
			Translation:   line.Translation,
			Notes:         notes,
			Model:         "precomputed",
			Version:       AnnotationVersion,
			CacheKey:      "demo:" + song.Slug,
		})
	}

	return builtSong{
		track: domain.Track{
			ID:         "demo:" + song.Slug,
			Provider:   domain.ProviderDemo,
			ProviderID: song.Slug,
			Title:      song.Title,
			Artist:     song.Artist,
			Album:      song.Album,
			DurationMs: int64(math.Round(duration * 1000)),
			AudioPath:  "",
		},
		lyrics: domain.Lyrics{
			ID:       "lrc_demo_" + song.Slug,
			TrackID:  "demo:" + song.Slug,
			Provider: "demo",
			Language: "en",
			Sync:     domain.SyncWord,
			Source:   "original score",
			License:  "Original material created for STEFANI (MIT)",
			Lines:    lines,
		},
		annot:   annot,
		spec:    spec,
		durance: duration,
	}
}

// fingerprint is used to detect a stale rendered file after a content change.
func fingerprint(song Song) string {
	h := sha1.New()
	h.Write([]byte(song.Slug))
	h.Write([]byte(strconv.FormatFloat(song.BPM, 'f', 2, 64)))
	for _, line := range song.Lines {
		h.Write([]byte(line.Text))
	}
	return hex.EncodeToString(h.Sum(nil))[:10]
}

func midiFreq(note int) float64 {
	return 440 * math.Pow(2, float64(note-69)/12)
}

func writeFileAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
