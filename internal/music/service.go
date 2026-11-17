// Package music contains the interchangeable music providers. The rest of the
// application talks to Catalog and Playback, never to a concrete service.
package music

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"

	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/storage"
)

// ErrUnavailable means a provider cannot serve requests right now (offline,
// feature flag off, upstream change). Callers surface it as an informational
// state, not as a crash.
var ErrUnavailable = errors.New("provider unavailable")

// Catalog is the read side: metadata only. Having a catalog entry says nothing
// about the right to play it.
type Catalog interface {
	Name() string
	Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error)
	BrowsePlaylist(ctx context.Context, playlistID string, limit int) ([]domain.Track, error)
}

// Service aggregates the configured providers. It is the single entry point the
// application layer uses for searching and importing.
type Service struct {
	db   *storage.DB
	yt   Catalog
	log  Logger
	ytOn func() bool
}

// Logger is the subset of slog.Logger the service needs.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
}

// NewService wires the providers. ytOn is read on every call so the feature
// flag takes effect without a restart.
func NewService(db *storage.DB, yt Catalog, log Logger, ytOn func() bool) *Service {
	if ytOn == nil {
		ytOn = func() bool { return false }
	}
	return &Service{db: db, yt: yt, log: log, ytOn: ytOn}
}

// Search queries local storage and, when enabled, YouTube Music. A failing
// external provider never hides local results.
func (s *Service) Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	query = strings.TrimSpace(query)
	if limit <= 0 || limit > 50 {
		limit = 25
	}
	local, err := s.searchLocal(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	if s.yt == nil || !s.ytOn() {
		return local, nil
	}
	remote, err := s.yt.Search(ctx, query, limit)
	if err != nil {
		// Degrade, do not fail: local results are still useful.
		s.log.Warn("youtube music search failed", "error", err)
		return local, nil
	}
	// A remote track that is already stored also matches the local query, so
	// the combined list must be deduplicated by provider identity.
	seen := make(map[string]bool, len(local)+len(remote))
	for _, r := range local {
		seen[providerKey(r.Track.Provider, r.Track.ProviderID)] = true
	}
	for _, r := range remote {
		key := providerKey(r.Track.Provider, r.Track.ProviderID)
		if seen[key] {
			continue
		}
		seen[key] = true
		local = append(local, r)
	}
	return local, nil
}

// providerKey identifies a track across providers.
func providerKey(provider domain.Provider, providerID string) string {
	return string(provider) + "\x00" + providerID
}

func (s *Service) searchLocal(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	tracks, err := s.db.SearchTracks(ctx, query, "", limit)
	if err != nil {
		return nil, err
	}
	out := make([]domain.SearchResult, 0, len(tracks))
	for _, t := range tracks {
		out = append(out, domain.SearchResult{
			Track:        t,
			ProviderName: string(t.Provider),
			Playable:     t.Playable(),
			Note:         playNote(t),
		})
	}
	return out, nil
}

func playNote(t domain.Track) string {
	if t.Playable() {
		return ""
	}
	if t.Provider == domain.ProviderYTMusic {
		return "Catalogo solamente: activa la reproduccion en linea para escucharla"
	}
	return "Sin audio local asociado"
}

// LocalAudioExtensions are the containers the webview can decode reliably.
var LocalAudioExtensions = []string{".mp3", ".m4a", ".aac", ".wav", ".flac", ".ogg", ".opus"}

// ImportFiles registers local audio files, reading their tags when present.
// Files that are not audio are reported back in skipped instead of failing the
// whole import.
func (s *Service) ImportFiles(ctx context.Context, paths []string) (imported []domain.Track, skipped []string, err error) {
	for _, path := range paths {
		if !IsAudioFile(path) {
			skipped = append(skipped, path)
			continue
		}
		track, err := s.importFile(ctx, path)
		if err != nil {
			s.log.Warn("cannot import audio file", "path", path, "error", err)
			skipped = append(skipped, path)
			continue
		}
		imported = append(imported, track)
	}
	return imported, skipped, nil
}

// IsAudioFile reports whether path has a supported audio extension.
func IsAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, want := range LocalAudioExtensions {
		if ext == want {
			return true
		}
	}
	return false
}

func (s *Service) importFile(ctx context.Context, path string) (domain.Track, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return domain.Track{}, fmt.Errorf("resolve path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return domain.Track{}, fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return domain.Track{}, fmt.Errorf("is a directory")
	}

	title := strings.TrimSuffix(filepath.Base(abs), filepath.Ext(abs))
	artist, album := "", ""
	if f, err := os.Open(abs); err == nil {
		if meta, err := tag.ReadFrom(f); err == nil {
			if v := strings.TrimSpace(meta.Title()); v != "" {
				title = v
			}
			artist = strings.TrimSpace(meta.Artist())
			album = strings.TrimSpace(meta.Album())
		}
		f.Close()
	}

	track := domain.Track{
		Provider:   domain.ProviderLocal,
		ProviderID: FileKey(abs),
		Title:      title,
		Artist:     artist,
		Album:      album,
		AudioPath:  abs,
	}
	id, err := s.db.SaveTrack(ctx, track)
	if err != nil {
		return domain.Track{}, err
	}
	track.ID = id
	return track, nil
}

// FileKey derives a stable provider id from an absolute path, so re-importing
// the same file updates the existing row.
func FileKey(abs string) string {
	replacer := strings.NewReplacer(`\`, "/", " ", "_")
	return "file_" + replacer.Replace(strings.ToLower(abs))
}

// UpdateDuration stores the duration reported by the audio engine.
func (s *Service) UpdateDuration(ctx context.Context, trackID string, durationMs int64) error {
	if durationMs <= 0 {
		return nil
	}
	return s.db.UpdateTrackDuration(ctx, trackID, durationMs)
}

// Resolve returns a track known to the database.
func (s *Service) Resolve(ctx context.Context, trackID string) (domain.Track, error) {
	return s.db.Track(ctx, trackID)
}
