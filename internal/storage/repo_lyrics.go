package storage

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// SaveLyrics writes the document for a track. IDs are content-addressed so a
// rebuild keeps annotation rows attached to unchanged lines and drops only the
// lines whose text actually changed.
func (d *DB) SaveLyrics(ctx context.Context, l domain.Lyrics) error {
	if l.TrackID == "" {
		return fmt.Errorf("lyrics need a track id")
	}
	if l.ID == "" {
		l.ID = "lrc_" + shortHash(l.TrackID)
	}
	for i := range l.Lines {
		if l.Lines[i].ID == "" {
			l.Lines[i].ID = fmt.Sprintf("%s-l%03d-%s", l.ID, i, shortHash(l.Lines[i].Text))
		}
	}

	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save lyrics: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO lyrics_documents (id, track_id, provider, language, sync_type, source_reference, license_metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider = excluded.provider,
			language = excluded.language,
			sync_type = excluded.sync_type,
			source_reference = excluded.source_reference,
			license_metadata = excluded.license_metadata`,
		l.ID, l.TrackID, l.Provider, def(l.Language, "en"), string(l.Sync),
		l.Source, l.License, nowMs()); err != nil {
		return fmt.Errorf("insert lyrics document: %w", err)
	}
	keep := make([]any, 0, len(l.Lines)+1)
	keep = append(keep, l.ID)
	for i, line := range l.Lines {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO lyrics_lines (id, document_id, position, start_ms, end_ms, text)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				position = excluded.position, start_ms = excluded.start_ms,
				end_ms = excluded.end_ms, text = excluded.text`,
			line.ID, l.ID, i, line.StartMs, line.EndMs, line.Text); err != nil {
			return fmt.Errorf("insert lyric line %d: %w", i, err)
		}
		keep = append(keep, line.ID)
		if _, err := tx.ExecContext(ctx, `DELETE FROM lyrics_words WHERE line_id = ?`, line.ID); err != nil {
			return fmt.Errorf("clear words of line %d: %w", i, err)
		}
		for j, w := range line.Words {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO lyrics_words (line_id, position, text, start_ms, end_ms, timing_source)
				VALUES (?, ?, ?, ?, ?, ?)`,
				line.ID, j, w.Text, w.StartMs, w.EndMs, def(w.TimingSrc, "provider")); err != nil {
				return fmt.Errorf("insert word %d of line %d: %w", j, i, err)
			}
		}
	}

	placeholders := make([]string, len(keep))
	for i := range keep {
		placeholders[i] = "?"
	}
	query := `DELETE FROM lyrics_lines WHERE document_id = ? AND id NOT IN (` + strings.Join(placeholders, ",") + `)`
	args := append([]any{l.ID}, keep...)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("drop stale lyric lines: %w", err)
	}
	return tx.Commit()
}

func shortHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}

// Lyrics returns the active document for a track.
func (d *DB) Lyrics(ctx context.Context, trackID string) (domain.Lyrics, error) {
	var l domain.Lyrics
	var created int64
	err := d.sql.QueryRowContext(ctx, `
		SELECT id, track_id, provider, language, sync_type, source_reference, license_metadata, created_at
		FROM lyrics_documents WHERE track_id = ? ORDER BY created_at DESC LIMIT 1`, trackID).
		Scan(&l.ID, &l.TrackID, &l.Provider, &l.Language, &l.Sync, &l.Source, &l.License, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Lyrics{}, ErrNotFound
	}
	if err != nil {
		return domain.Lyrics{}, fmt.Errorf("load lyrics: %w", err)
	}

	rows, err := d.sql.QueryContext(ctx, `
		SELECT id, position, start_ms, end_ms, text FROM lyrics_lines
		WHERE document_id = ? ORDER BY position`, l.ID)
	if err != nil {
		return domain.Lyrics{}, fmt.Errorf("load lyric lines: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var line domain.Line
		if err := rows.Scan(&line.ID, &line.Index, &line.StartMs, &line.EndMs, &line.Text); err != nil {
			return domain.Lyrics{}, err
		}
		l.Lines = append(l.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return domain.Lyrics{}, err
	}
	if err := rows.Close(); err != nil {
		return domain.Lyrics{}, err
	}

	wordRows, err := d.sql.QueryContext(ctx, `
		SELECT w.line_id, w.text, w.start_ms, w.end_ms, w.timing_source
		FROM lyrics_words w JOIN lyrics_lines l ON l.id = w.line_id
		WHERE l.document_id = ? ORDER BY w.line_id, w.position`, l.ID)
	if err != nil {
		return domain.Lyrics{}, fmt.Errorf("load lyric words: %w", err)
	}
	defer wordRows.Close()
	byLine := make(map[string][]domain.Word)
	for wordRows.Next() {
		var lineID string
		var w domain.Word
		if err := wordRows.Scan(&lineID, &w.Text, &w.StartMs, &w.EndMs, &w.TimingSrc); err != nil {
			return domain.Lyrics{}, err
		}
		byLine[lineID] = append(byLine[lineID], w)
	}
	if err := wordRows.Err(); err != nil {
		return domain.Lyrics{}, err
	}
	for i := range l.Lines {
		l.Lines[i].Words = byLine[l.Lines[i].ID]
	}
	return l, nil
}

// SaveAnnotation upserts one annotation for a line and accent.
func (d *DB) SaveAnnotation(ctx context.Context, a domain.Annotation) error {
	if a.Accent == "" {
		a.Accent = "en-US"
	}
	notes, err := json.Marshal(a.Notes)
	if err != nil {
		return fmt.Errorf("encode notes: %w", err)
	}
	_, err = d.sql.ExecContext(ctx, `
		INSERT INTO phonetic_annotations
			(id, line_id, accent, ipa, spanish_approx, translation, notes_json, model, generator_version, cache_key, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(line_id, accent) DO UPDATE SET
			ipa = excluded.ipa, spanish_approx = excluded.spanish_approx,
			translation = excluded.translation, notes_json = excluded.notes_json,
			model = excluded.model, generator_version = excluded.generator_version,
			cache_key = excluded.cache_key, created_at = excluded.created_at`,
		newID("ann"), a.LineID, a.Accent, a.IPA, a.SpanishApprox, a.Translation,
		string(notes), a.Model, a.Version, a.CacheKey, nowMs())
	if err != nil {
		return fmt.Errorf("save annotation: %w", err)
	}
	return nil
}

// AnnotationsByTrack loads every annotation of a track's current document.
func (d *DB) AnnotationsByTrack(ctx context.Context, trackID string) ([]domain.Annotation, error) {
	rows, err := d.sql.QueryContext(ctx, `
		SELECT a.line_id, a.accent, a.ipa, a.spanish_approx, a.translation, a.notes_json,
		       a.model, a.generator_version, a.cache_key, a.created_at
		FROM phonetic_annotations a
		JOIN lyrics_lines l ON l.id = a.line_id
		JOIN lyrics_documents doc ON doc.id = l.document_id
		WHERE doc.track_id = ?`, trackID)
	if err != nil {
		return nil, fmt.Errorf("load annotations: %w", err)
	}
	defer rows.Close()
	var out []domain.Annotation
	for rows.Next() {
		var a domain.Annotation
		var notes string
		var created int64
		if err := rows.Scan(&a.LineID, &a.Accent, &a.IPA, &a.SpanishApprox, &a.Translation,
			&notes, &a.Model, &a.Version, &a.CacheKey, &created); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(notes), &a.Notes); err != nil {
			a.Notes = nil
		}
		a.CreatedAt = msToTime(created)
		out = append(out, a)
	}
	return out, rows.Err()
}

// AnnotationForLine returns one annotation, if present.
func (d *DB) AnnotationForLine(ctx context.Context, lineID, accent string) (domain.Annotation, error) {
	var a domain.Annotation
	var notes string
	var created int64
	err := d.sql.QueryRowContext(ctx, `
		SELECT line_id, accent, ipa, spanish_approx, translation, notes_json, model,
		       generator_version, cache_key, created_at
		FROM phonetic_annotations WHERE line_id = ? AND accent = ?`, lineID, accent).
		Scan(&a.LineID, &a.Accent, &a.IPA, &a.SpanishApprox, &a.Translation, &notes,
			&a.Model, &a.Version, &a.CacheKey, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Annotation{}, ErrNotFound
	}
	if err != nil {
		return domain.Annotation{}, fmt.Errorf("load annotation: %w", err)
	}
	_ = json.Unmarshal([]byte(notes), &a.Notes)
	a.CreatedAt = msToTime(created)
	return a, nil
}

// DeleteAnnotationsForTrack drops generated study data, keeping the lyrics.
func (d *DB) DeleteAnnotationsForTrack(ctx context.Context, trackID string) error {
	_, err := d.sql.ExecContext(ctx, `
		DELETE FROM phonetic_annotations WHERE line_id IN (
			SELECT l.id FROM lyrics_lines l
			JOIN lyrics_documents doc ON doc.id = l.document_id
			WHERE doc.track_id = ?)`, trackID)
	if err != nil {
		return fmt.Errorf("delete annotations: %w", err)
	}
	return nil
}

func def(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// nowTime is used by repositories that expose creation timestamps to callers.
func nowTime() time.Time { return time.Now().UTC() }
