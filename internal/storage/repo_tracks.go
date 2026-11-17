package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// ErrNotFound is returned when a lookup finds no row.
var ErrNotFound = errors.New("not found")

// SaveTrack inserts or updates a track and returns its canonical ID.
func (d *DB) SaveTrack(ctx context.Context, t domain.Track) (string, error) {
	if t.ID == "" {
		t.ID = string(t.Provider) + ":" + t.ProviderID
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO tracks (id, provider, provider_track_id, title, artist, album,
		                    duration_ms, artwork_url, external_url, audio_path, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			artist = excluded.artist,
			album = excluded.album,
			duration_ms = excluded.duration_ms,
			artwork_url = excluded.artwork_url,
			external_url = excluded.external_url,
			audio_path = excluded.audio_path`,
		t.ID, string(t.Provider), t.ProviderID, t.Title, t.Artist, t.Album,
		t.DurationMs, t.ArtworkURL, t.ExternalURL, t.AudioPath, t.CreatedAt.UnixMilli())
	if err != nil {
		return "", fmt.Errorf("save track: %w", err)
	}
	return t.ID, nil
}

func scanTrack(row interface{ Scan(...any) error }) (domain.Track, error) {
	var t domain.Track
	var provider string
	var created int64
	err := row.Scan(&t.ID, &provider, &t.ProviderID, &t.Title, &t.Artist, &t.Album,
		&t.DurationMs, &t.ArtworkURL, &t.ExternalURL, &t.AudioPath, &created)
	if err != nil {
		return domain.Track{}, err
	}
	t.Provider = domain.Provider(provider)
	t.CreatedAt = msToTime(created)
	return t, nil
}

const trackColumns = `id, provider, provider_track_id, title, artist, album, duration_ms,
	artwork_url, external_url, audio_path, created_at`

// Track fetches one track by ID.
func (d *DB) Track(ctx context.Context, id string) (domain.Track, error) {
	row := d.sql.QueryRowContext(ctx, `SELECT `+trackColumns+` FROM tracks WHERE id = ?`, id)
	t, err := scanTrack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Track{}, ErrNotFound
	}
	if err != nil {
		return domain.Track{}, fmt.Errorf("load track: %w", err)
	}
	return t, nil
}

// TrackByProviderID looks up a track by provider and external identifier.
func (d *DB) TrackByProviderID(ctx context.Context, p domain.Provider, providerID string) (domain.Track, error) {
	row := d.sql.QueryRowContext(ctx,
		`SELECT `+trackColumns+` FROM tracks WHERE provider = ? AND provider_track_id = ?`,
		string(p), providerID)
	t, err := scanTrack(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Track{}, ErrNotFound
	}
	if err != nil {
		return domain.Track{}, fmt.Errorf("load track by provider id: %w", err)
	}
	return t, nil
}

// TracksByIDs loads many tracks, preserving the order of ids and skipping
// unknown ids.
func (d *DB) TracksByIDs(ctx context.Context, ids []string) ([]domain.Track, error) {
	out := make([]domain.Track, 0, len(ids))
	for _, id := range ids {
		t, err := d.Track(ctx, id)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// LibraryTrackIDs returns saved track IDs, newest first.
func (d *DB) LibraryTrackIDs(ctx context.Context) ([]string, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT track_id FROM library_tracks ORDER BY saved_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list library: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Favorites returns favourite track IDs, newest first.
func (d *DB) Favorites(ctx context.Context) ([]string, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT track_id FROM library_tracks WHERE is_favorite = 1 ORDER BY saved_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// SaveToLibrary adds a track to the library, keeping its favourite flag.
func (d *DB) SaveToLibrary(ctx context.Context, trackID string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO library_tracks (track_id, is_favorite, saved_at) VALUES (?, 0, ?)
		ON CONFLICT(track_id) DO NOTHING`, trackID, nowMs())
	if err != nil {
		return fmt.Errorf("save to library: %w", err)
	}
	return nil
}

// RemoveFromLibrary deletes a track from the library.
func (d *DB) RemoveFromLibrary(ctx context.Context, trackID string) error {
	_, err := d.sql.ExecContext(ctx, `DELETE FROM library_tracks WHERE track_id = ?`, trackID)
	if err != nil {
		return fmt.Errorf("remove from library: %w", err)
	}
	return nil
}

// IsFavorite reports the favourite flag for a library track.
func (d *DB) IsFavorite(ctx context.Context, trackID string) (bool, bool) {
	var fav bool
	err := d.sql.QueryRowContext(ctx,
		`SELECT is_favorite FROM library_tracks WHERE track_id = ?`, trackID).Scan(&fav)
	if err != nil {
		return false, false
	}
	return fav, true
}

// SetFavorite flags or unflags a track, adding it to the library when flagged.
func (d *DB) SetFavorite(ctx context.Context, trackID string, favorite bool) error {
	value := 0
	if favorite {
		value = 1
	}
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO library_tracks (track_id, is_favorite, saved_at) VALUES (?, ?, ?)
		ON CONFLICT(track_id) DO UPDATE SET is_favorite = excluded.is_favorite`,
		trackID, value, nowMs())
	if err != nil {
		return fmt.Errorf("set favorite: %w", err)
	}
	return nil
}

// RecordPlayback appends a history row.
func (d *DB) RecordPlayback(ctx context.Context, trackID string, listenedMs int64) error {
	_, err := d.sql.ExecContext(ctx,
		`INSERT INTO playback_history (track_id, played_at, duration_listened_ms) VALUES (?, ?, ?)`,
		trackID, nowMs(), listenedMs)
	if err != nil {
		return fmt.Errorf("record playback: %w", err)
	}
	return nil
}

// History returns recent playback entries joined with track metadata.
func (d *DB) History(ctx context.Context, limit int) ([]domain.HistoryEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.sql.QueryContext(ctx, `
		SELECT h.id, h.track_id, h.played_at, h.duration_listened_ms, t.title, t.artist
		FROM playback_history h JOIN tracks t ON t.id = h.track_id
		ORDER BY h.played_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	defer rows.Close()
	out := make([]domain.HistoryEntry, 0, limit)
	for rows.Next() {
		var e domain.HistoryEntry
		var played int64
		if err := rows.Scan(&e.ID, &e.TrackID, &played, &e.ListenedMs, &e.TrackTitle, &e.TrackArtist); err != nil {
			return nil, err
		}
		e.PlayedAt = msToTime(played)
		out = append(out, e)
	}
	return out, rows.Err()
}

// RecentlyPlayed returns distinct track IDs from history, newest first.
func (d *DB) RecentlyPlayed(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 12
	}
	rows, err := d.sql.QueryContext(ctx, `
		SELECT track_id, MAX(played_at) AS last_played
		FROM playback_history GROUP BY track_id
		ORDER BY last_played DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list recently played: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		var last int64
		if err := rows.Scan(&id, &last); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// AllTracks returns every stored track, used by local search.
func (d *DB) AllTracks(ctx context.Context) ([]domain.Track, error) {
	rows, err := d.sql.QueryContext(ctx, `SELECT `+trackColumns+` FROM tracks ORDER BY title COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("list tracks: %w", err)
	}
	defer rows.Close()
	var out []domain.Track
	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
