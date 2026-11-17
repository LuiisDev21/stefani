package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// SearchTracks matches title or artist, optionally restricted to a provider.
func (d *DB) SearchTracks(ctx context.Context, query, provider string, limit int) ([]domain.Track, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	like := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	// Online rows cached before the search asked for the Songs tab can be
	// podcasts or reaction videos: no artist, no duration. They are kept (the
	// lyric may be cached) but never offered as a search result again.
	rows, err := d.sql.QueryContext(ctx, `
		SELECT `+trackColumns+` FROM tracks
		WHERE (LOWER(title) LIKE ? OR LOWER(artist) LIKE ? OR LOWER(album) LIKE ?)
		  AND (? = '' OR provider = ?)
		  AND NOT (provider = 'ytmusic' AND COALESCE(artist, '') = '' AND COALESCE(duration_ms, 0) = 0)
		ORDER BY title COLLATE NOCASE LIMIT ?`, like, like, like, provider, provider, limit)
	if err != nil {
		return nil, fmt.Errorf("search tracks: %w", err)
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

// TracksByProvider lists every track of a provider.
func (d *DB) TracksByProvider(ctx context.Context, provider domain.Provider) ([]domain.Track, error) {
	rows, err := d.sql.QueryContext(ctx,
		`SELECT `+trackColumns+` FROM tracks WHERE provider = ? ORDER BY created_at`, string(provider))
	if err != nil {
		return nil, fmt.Errorf("list provider tracks: %w", err)
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

// UpdateTrackDuration stores the duration reported by the player.
func (d *DB) UpdateTrackDuration(ctx context.Context, trackID string, durationMs int64) error {
	_, err := d.sql.ExecContext(ctx,
		`UPDATE tracks SET duration_ms = ? WHERE id = ? AND duration_ms = 0`, durationMs, trackID)
	if err != nil {
		return fmt.Errorf("update track duration: %w", err)
	}
	return nil
}

// CountLibrary returns how many tracks are saved.
func (d *DB) CountLibrary(ctx context.Context) (int, error) {
	var n int
	err := d.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM library_tracks`).Scan(&n)
	return n, err
}

// CountFavorites returns how many saved tracks are favourites.
func (d *DB) CountFavorites(ctx context.Context) (int, error) {
	var n int
	err := d.sql.QueryRowContext(ctx, `SELECT COUNT(*) FROM library_tracks WHERE is_favorite = 1`).Scan(&n)
	return n, err
}

// LibraryTrack returns a saved track by ID, or ErrNotFound.
func (d *DB) LibraryTrack(ctx context.Context, trackID string) (domain.Track, error) {
	var exists int
	err := d.sql.QueryRowContext(ctx, `SELECT 1 FROM library_tracks WHERE track_id = ?`, trackID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Track{}, ErrNotFound
	}
	if err != nil {
		return domain.Track{}, fmt.Errorf("check library: %w", err)
	}
	return d.Track(ctx, trackID)
}
