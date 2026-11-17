package lyrics

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/storage"
)

// Logger is the logging surface the service needs.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

// Service resolves lyrics for a track: local sidecar file first, then LRCLIB,
// always caching the result in SQLite.
type Service struct {
	db  *storage.DB
	lib *LRCLIB
	log Logger
}

// NewService wires the lyric service.
func NewService(db *storage.DB, lib *LRCLIB, log Logger) *Service {
	return &Service{db: db, lib: lib, log: log}
}

// ForTrack returns lyrics for a track, fetching them when not cached. refresh
// forces a provider lookup even if a document exists.
func (s *Service) ForTrack(ctx context.Context, track domain.Track, refresh bool) (domain.Lyrics, error) {
	if !refresh {
		if cached, err := s.db.Lyrics(ctx, track.ID); err == nil && len(cached.Lines) > 0 {
			return cached, nil
		}
	}

	if path := sidecarPath(track.AudioPath); path != "" {
		doc, err := s.importPath(ctx, track, path)
		if err == nil {
			return doc, nil
		}
		s.log.Warn("sidecar lyrics could not be read", "path", path, "error", err)
	}

	if s.lib == nil {
		return domain.Lyrics{}, ErrNoLyrics
	}
	remote, err := s.lib.Lookup(ctx, track)
	if err != nil {
		if errors.Is(err, ErrNoLyrics) {
			return domain.Lyrics{}, ErrNoLyrics
		}
		return domain.Lyrics{}, err
	}
	if remote == nil {
		return domain.Lyrics{}, ErrNoLyrics
	}
	raw := remote.SyncedLyrics
	source := "lrclib:synced"
	if raw == "" {
		raw = remote.PlainLyrics
		source = "lrclib:plain"
	}
	parsed := Parse(raw)
	if len(parsed.Lines) == 0 {
		return domain.Lyrics{}, ErrNoLyrics
	}
	doc := domain.Lyrics{
		TrackID:  track.ID,
		Provider: "lrclib",
		Language: "en",
		Sync:     parsed.Sync,
		Source:   fmt.Sprintf("%s#%d", source, remote.ID),
		License:  "LRCLIB community database, cached locally for private study",
		Lines:    parsed.Lines,
	}
	if err := s.db.SaveLyrics(ctx, doc); err != nil {
		return domain.Lyrics{}, err
	}
	saved, err := s.db.Lyrics(ctx, track.ID)
	if err != nil {
		return doc, nil
	}
	return saved, nil
}

// ImportFile stores a lyric file chosen by the user for a track.
func (s *Service) ImportFile(ctx context.Context, track domain.Track, path string) (domain.Lyrics, error) {
	return s.importPath(ctx, track, path)
}

func (s *Service) importPath(ctx context.Context, track domain.Track, path string) (domain.Lyrics, error) {
	info, err := os.Stat(path)
	if err != nil {
		return domain.Lyrics{}, fmt.Errorf("read lyrics file: %w", err)
	}
	if info.Size() > 4<<20 {
		return domain.Lyrics{}, fmt.Errorf("lyrics file is too large (limit 4 MB)")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return domain.Lyrics{}, fmt.Errorf("read lyrics file: %w", err)
	}
	parsed := Parse(string(raw))
	if len(parsed.Lines) == 0 {
		return domain.Lyrics{}, fmt.Errorf("the file contains no lyric lines")
	}
	doc := domain.Lyrics{
		TrackID:  track.ID,
		Provider: "file",
		Language: "en",
		Sync:     parsed.Sync,
		Source:   filepath.Base(path),
		License:  "provided by the user",
		Lines:    parsed.Lines,
	}
	if err := s.db.SaveLyrics(ctx, doc); err != nil {
		return domain.Lyrics{}, err
	}
	return s.db.Lyrics(ctx, track.ID)
}

// Delete removes the cached document for a track.
func (s *Service) Delete(ctx context.Context, trackID string) error {
	return s.db.DeleteLyrics(ctx, trackID)
}

// sidecarPath returns the .lrc next to an audio file, if it exists.
func sidecarPath(audioPath string) string {
	if audioPath == "" {
		return ""
	}
	candidate := strings.TrimSuffix(audioPath, filepath.Ext(audioPath)) + ".lrc"
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}
	return ""
}
