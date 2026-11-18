package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// CreatePlaylist inserts a playlist and returns it with its generated ID.
func (d *DB) CreatePlaylist(ctx context.Context, p domain.Playlist) (domain.Playlist, error) {
	if p.Name == "" {
		return domain.Playlist{}, fmt.Errorf("playlist name is required")
	}
	if p.ID == "" {
		p.ID = newID("pl")
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	if p.Origin == "" {
		p.Origin = "local"
	}
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO playlists (id, name, description, artwork, origin, external_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Description, p.Artwork, p.Origin, p.ExternalID,
		p.CreatedAt.UnixMilli(), p.UpdatedAt.UnixMilli())
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("create playlist: %w", err)
	}
	return p, nil
}

// UpdatePlaylist renames or re-describes a playlist.
func (d *DB) UpdatePlaylist(ctx context.Context, id, name, description string) error {
	if name == "" {
		return fmt.Errorf("playlist name is required")
	}
	res, err := d.sql.ExecContext(ctx,
		`UPDATE playlists SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		name, description, nowMs(), id)
	if err != nil {
		return fmt.Errorf("update playlist: %w", err)
	}
	return requireAffected(res, "playlist")
}

// DeletePlaylist removes a playlist and its items (cascade).
func (d *DB) DeletePlaylist(ctx context.Context, id string) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM playlists WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete playlist: %w", err)
	}
	return requireAffected(res, "playlist")
}

// Playlists lists playlists with their item counts.
func (d *DB) Playlists(ctx context.Context) ([]domain.Playlist, error) {
	rows, err := d.sql.QueryContext(ctx, `
		SELECT p.id, p.name, p.description, COALESCE(NULLIF(p.artwork, ''), (
		  SELECT t.artwork_url FROM playlist_items pi
		  JOIN tracks t ON t.id = pi.track_id
		  WHERE pi.playlist_id = p.id AND t.artwork_url <> ''
		  ORDER BY pi.position LIMIT 1
		), ''), p.origin, p.external_id,
		       p.created_at, p.updated_at, COUNT(i.id)
		FROM playlists p LEFT JOIN playlist_items i ON i.playlist_id = p.id
		GROUP BY p.id ORDER BY p.updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list playlists: %w", err)
	}
	defer rows.Close()
	var out []domain.Playlist
	for rows.Next() {
		var p domain.Playlist
		var created, updated int64
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Artwork, &p.Origin,
			&p.ExternalID, &created, &updated, &p.TrackCount); err != nil {
			return nil, err
		}
		p.CreatedAt = msToTime(created)
		p.UpdatedAt = msToTime(updated)
		out = append(out, p)
	}
	return out, rows.Err()
}

// Playlist fetches one playlist.
func (d *DB) Playlist(ctx context.Context, id string) (domain.Playlist, error) {
	var p domain.Playlist
	var created, updated int64
	err := d.sql.QueryRowContext(ctx, `
		SELECT p.id, p.name, p.description, COALESCE(NULLIF(p.artwork, ''), (
		  SELECT t.artwork_url FROM playlist_items pi
		  JOIN tracks t ON t.id = pi.track_id
		  WHERE pi.playlist_id = p.id AND t.artwork_url <> ''
		  ORDER BY pi.position LIMIT 1
		), ''), p.origin, p.external_id,
		       p.created_at, p.updated_at, (SELECT COUNT(*) FROM playlist_items i WHERE i.playlist_id = p.id)
		FROM playlists p WHERE p.id = ?`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Artwork, &p.Origin, &p.ExternalID, &created, &updated, &p.TrackCount)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Playlist{}, ErrNotFound
	}
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("load playlist: %w", err)
	}
	p.CreatedAt = msToTime(created)
	p.UpdatedAt = msToTime(updated)
	return p, nil
}

// PlaylistItems lists the ordered tracks of a playlist.
func (d *DB) PlaylistItems(ctx context.Context, playlistID string) ([]domain.PlaylistItem, error) {
	rows, err := d.sql.QueryContext(ctx, `
		SELECT i.position, t.id, t.provider, t.provider_track_id, t.title, t.artist, t.album,
		       t.duration_ms, t.artwork_url, t.external_url, t.audio_path, t.created_at
		FROM playlist_items i JOIN tracks t ON t.id = i.track_id
		WHERE i.playlist_id = ? ORDER BY i.position`, playlistID)
	if err != nil {
		return nil, fmt.Errorf("list playlist items: %w", err)
	}
	defer rows.Close()
	var out []domain.PlaylistItem
	for rows.Next() {
		var item domain.PlaylistItem
		var provider string
		var created int64
		if err := rows.Scan(&item.Position, &item.Track.ID, &provider, &item.Track.ProviderID,
			&item.Track.Title, &item.Track.Artist, &item.Track.Album, &item.Track.DurationMs,
			&item.Track.ArtworkURL, &item.Track.ExternalURL, &item.Track.AudioPath, &created); err != nil {
			return nil, err
		}
		item.Track.Provider = domain.Provider(provider)
		item.Track.CreatedAt = msToTime(created)
		out = append(out, item)
	}
	return out, rows.Err()
}

// AddToPlaylist appends a track. Duplicate tracks are allowed: a playlist is a
// sequence, not a set.
func (d *DB) AddToPlaylist(ctx context.Context, playlistID, trackID string) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin add to playlist: %w", err)
	}
	defer tx.Rollback()

	var next int
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(position), -1) + 1 FROM playlist_items WHERE playlist_id = ?`,
		playlistID).Scan(&next); err != nil {
		return fmt.Errorf("next playlist position: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO playlist_items (playlist_id, track_id, position, added_at) VALUES (?, ?, ?, ?)`,
		playlistID, trackID, next, nowMs()); err != nil {
		return fmt.Errorf("add playlist item: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE playlists SET updated_at = ? WHERE id = ?`, nowMs(), playlistID); err != nil {
		return fmt.Errorf("touch playlist: %w", err)
	}
	return tx.Commit()
}

// RemovePlaylistItem removes the item at position and closes the gap.
func (d *DB) RemovePlaylistItem(ctx context.Context, playlistID string, position int) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin remove playlist item: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`DELETE FROM playlist_items WHERE playlist_id = ? AND position = ?`, playlistID, position)
	if err != nil {
		return fmt.Errorf("delete playlist item: %w", err)
	}
	if err := requireAffected(res, "playlist item"); err != nil {
		return err
	}
	if err := compactPositions(ctx, tx, playlistID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE playlists SET updated_at = ? WHERE id = ?`, nowMs(), playlistID); err != nil {
		return fmt.Errorf("touch playlist: %w", err)
	}
	return tx.Commit()
}

// MovePlaylistItem moves the item at from to index to, shifting the rest.
func (d *DB) MovePlaylistItem(ctx context.Context, playlistID string, from, to int) error {
	if from == to {
		return nil
	}
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin move playlist item: %w", err)
	}
	defer tx.Rollback()

	var trackID string
	if err := tx.QueryRowContext(ctx,
		`SELECT track_id FROM playlist_items WHERE playlist_id = ? AND position = ?`,
		playlistID, from).Scan(&trackID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("load moving item: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM playlist_items WHERE playlist_id = ? AND position = ?`, playlistID, from); err != nil {
		return fmt.Errorf("detach moving item: %w", err)
	}
	if err := compactPositions(ctx, tx, playlistID); err != nil {
		return err
	}

	var count int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM playlist_items WHERE playlist_id = ?`, playlistID).Scan(&count); err != nil {
		return fmt.Errorf("count playlist items: %w", err)
	}
	if to < 0 {
		to = 0
	}
	if to > count {
		to = count
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE playlist_items SET position = position + 1 WHERE playlist_id = ? AND position >= ?`,
		playlistID, to); err != nil {
		return fmt.Errorf("shift playlist items: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO playlist_items (playlist_id, track_id, position, added_at) VALUES (?, ?, ?, ?)`,
		playlistID, trackID, to, nowMs()); err != nil {
		return fmt.Errorf("insert moved item: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE playlists SET updated_at = ? WHERE id = ?`, nowMs(), playlistID); err != nil {
		return fmt.Errorf("touch playlist: %w", err)
	}
	return tx.Commit()
}

func compactPositions(ctx context.Context, tx *sql.Tx, playlistID string) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT id FROM playlist_items WHERE playlist_id = ? ORDER BY position, id`, playlistID)
	if err != nil {
		return fmt.Errorf("read playlist order: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE playlist_items SET position = ? WHERE id = ?`, i, id); err != nil {
			return fmt.Errorf("compact playlist positions: %w", err)
		}
	}
	return nil
}

func requireAffected(res sql.Result, entity string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check %s update: %w", entity, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
