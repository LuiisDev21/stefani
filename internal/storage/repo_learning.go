package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// SaveVocabulary inserts or updates a vocabulary entry. The unique index on
// word makes saving the same word twice an update, not a duplicate.
func (d *DB) SaveVocabulary(ctx context.Context, w domain.WordEntry) (domain.WordEntry, error) {
	if w.Word == "" {
		return domain.WordEntry{}, fmt.Errorf("word is required")
	}
	if w.ID == "" {
		w.ID = newID("voc")
	}
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now().UTC()
	}
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO saved_vocabulary
			(id, word, lemma, part_of_speech, meaning, ipa, translation, example, source, track_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(word) DO UPDATE SET
			lemma = excluded.lemma, part_of_speech = excluded.part_of_speech,
			meaning = excluded.meaning, ipa = excluded.ipa, translation = excluded.translation,
			example = excluded.example, source = excluded.source, track_id = excluded.track_id`,
		w.ID, w.Word, w.Lemma, w.PartOfSpeech, w.Meaning, w.IPA, w.Translation,
		w.Example, def(w.Source, "ai"), w.TrackID, w.CreatedAt.UnixMilli())
	if err != nil {
		return domain.WordEntry{}, fmt.Errorf("save vocabulary: %w", err)
	}
	return w, nil
}

// Vocabulary lists saved words, newest first.
func (d *DB) Vocabulary(ctx context.Context) ([]domain.WordEntry, error) {
	rows, err := d.sql.QueryContext(ctx, `
		SELECT id, word, lemma, part_of_speech, meaning, ipa, translation, example, source, track_id, created_at
		FROM saved_vocabulary ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list vocabulary: %w", err)
	}
	defer rows.Close()
	var out []domain.WordEntry
	for rows.Next() {
		var w domain.WordEntry
		var created int64
		if err := rows.Scan(&w.ID, &w.Word, &w.Lemma, &w.PartOfSpeech, &w.Meaning, &w.IPA,
			&w.Translation, &w.Example, &w.Source, &w.TrackID, &created); err != nil {
			return nil, err
		}
		w.CreatedAt = msToTime(created)
		out = append(out, w)
	}
	return out, rows.Err()
}

// DeleteVocabulary removes a saved word.
func (d *DB) DeleteVocabulary(ctx context.Context, id string) error {
	res, err := d.sql.ExecContext(ctx, `DELETE FROM saved_vocabulary WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete vocabulary: %w", err)
	}
	return requireAffected(res, "vocabulary")
}

// RecordStudy adds study time and practiced lines to a track's progress.
func (d *DB) RecordStudy(ctx context.Context, trackID string, seconds int64, lines int) error {
	if seconds < 0 {
		seconds = 0
	}
	if lines < 0 {
		lines = 0
	}
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO learning_progress (track_id, lines_practiced, study_seconds, last_studied_at, status)
		VALUES (?, ?, ?, ?, 'learning')
		ON CONFLICT(track_id) DO UPDATE SET
			lines_practiced = learning_progress.lines_practiced + excluded.lines_practiced,
			study_seconds = learning_progress.study_seconds + excluded.study_seconds,
			last_studied_at = excluded.last_studied_at,
			status = CASE WHEN learning_progress.status = 'new' THEN 'learning' ELSE learning_progress.status END`,
		trackID, lines, seconds, nowMs())
	if err != nil {
		return fmt.Errorf("record study: %w", err)
	}
	return nil
}

// SetTrackStatus sets the learning status of a track.
func (d *DB) SetTrackStatus(ctx context.Context, trackID, status string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO learning_progress (track_id, lines_practiced, study_seconds, last_studied_at, status)
		VALUES (?, 0, 0, ?, ?)
		ON CONFLICT(track_id) DO UPDATE SET status = excluded.status, last_studied_at = excluded.last_studied_at`,
		trackID, nowMs(), status)
	if err != nil {
		return fmt.Errorf("set track status: %w", err)
	}
	return nil
}

// ProgressForTrack returns progress for one track.
func (d *DB) ProgressForTrack(ctx context.Context, trackID string) (domain.Progress, error) {
	var p domain.Progress
	var last int64
	err := d.sql.QueryRowContext(ctx, `
		SELECT track_id, lines_practiced, study_seconds, last_studied_at, status
		FROM learning_progress WHERE track_id = ?`, trackID).
		Scan(&p.TrackID, &p.LinesPracticed, &p.StudySeconds, &last, &p.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Progress{TrackID: trackID, Status: "new"}, nil
	}
	if err != nil {
		return domain.Progress{}, fmt.Errorf("load progress: %w", err)
	}
	p.LastStudiedAt = msToTime(last)
	return p, nil
}

// AllProgress lists progress rows, most recently studied first.
func (d *DB) AllProgress(ctx context.Context) ([]domain.Progress, error) {
	rows, err := d.sql.QueryContext(ctx, `
		SELECT track_id, lines_practiced, study_seconds, last_studied_at, status
		FROM learning_progress ORDER BY last_studied_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}
	defer rows.Close()
	var out []domain.Progress
	for rows.Next() {
		var p domain.Progress
		var last int64
		if err := rows.Scan(&p.TrackID, &p.LinesPracticed, &p.StudySeconds, &last, &p.Status); err != nil {
			return nil, err
		}
		p.LastStudiedAt = msToTime(last)
		out = append(out, p)
	}
	return out, rows.Err()
}

// Setting reads a raw settings value.
func (d *DB) Setting(ctx context.Context, key string) (string, error) {
	var value string
	err := d.sql.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read setting %q: %w", key, err)
	}
	return value, nil
}

// SetSetting writes a raw settings value.
func (d *DB) SetSetting(ctx context.Context, key, value string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	if err != nil {
		return fmt.Errorf("write setting %q: %w", key, err)
	}
	return nil
}

// SetProviderConnection records the state of an external provider.
func (d *DB) SetProviderConnection(ctx context.Context, provider, state, metadata string) error {
	_, err := d.sql.ExecContext(ctx, `
		INSERT INTO provider_connections (provider, connection_state, metadata, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(provider) DO UPDATE SET
			connection_state = excluded.connection_state,
			metadata = excluded.metadata,
			updated_at = excluded.updated_at`, provider, state, metadata, nowMs())
	if err != nil {
		return fmt.Errorf("save provider connection: %w", err)
	}
	return nil
}

// ProviderConnection reads a provider's connection state.
func (d *DB) ProviderConnection(ctx context.Context, provider string) (state string, metadata string, err error) {
	err = d.sql.QueryRowContext(ctx,
		`SELECT connection_state, metadata FROM provider_connections WHERE provider = ?`, provider).
		Scan(&state, &metadata)
	if errors.Is(err, sql.ErrNoRows) {
		return "disconnected", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("read provider connection: %w", err)
	}
	return state, metadata, nil
}
