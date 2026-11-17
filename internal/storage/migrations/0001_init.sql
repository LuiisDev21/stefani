-- 0001_init.sql
-- Base schema for STEFANI. All timestamps are Unix milliseconds unless the
-- column name says otherwise.

CREATE TABLE tracks (
    id                TEXT PRIMARY KEY,
    provider          TEXT NOT NULL,
    provider_track_id TEXT NOT NULL,
    title             TEXT NOT NULL,
    artist            TEXT NOT NULL DEFAULT '',
    album             TEXT NOT NULL DEFAULT '',
    duration_ms       INTEGER NOT NULL DEFAULT 0,
    artwork_url       TEXT NOT NULL DEFAULT '',
    external_url      TEXT NOT NULL DEFAULT '',
    audio_path        TEXT NOT NULL DEFAULT '',
    created_at        INTEGER NOT NULL,
    UNIQUE (provider, provider_track_id)
);

CREATE INDEX idx_tracks_title  ON tracks(title);
CREATE INDEX idx_tracks_artist ON tracks(artist);

CREATE TABLE playlists (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    artwork     TEXT NOT NULL DEFAULT '',
    origin      TEXT NOT NULL DEFAULT 'local',
    external_id TEXT NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE playlist_items (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    playlist_id TEXT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    track_id    TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    position    INTEGER NOT NULL,
    added_at    INTEGER NOT NULL
);

CREATE INDEX idx_playlist_items ON playlist_items(playlist_id, position);

CREATE TABLE library_tracks (
    track_id    TEXT PRIMARY KEY REFERENCES tracks(id) ON DELETE CASCADE,
    is_favorite INTEGER NOT NULL DEFAULT 0,
    saved_at    INTEGER NOT NULL
);

CREATE TABLE lyrics_documents (
    id               TEXT PRIMARY KEY,
    track_id         TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    provider         TEXT NOT NULL,
    language         TEXT NOT NULL DEFAULT 'en',
    sync_type        TEXT NOT NULL,
    source_reference TEXT NOT NULL DEFAULT '',
    license_metadata TEXT NOT NULL DEFAULT '',
    created_at       INTEGER NOT NULL
);

CREATE INDEX idx_lyrics_documents_track ON lyrics_documents(track_id);

CREATE TABLE lyrics_lines (
    id          TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES lyrics_documents(id) ON DELETE CASCADE,
    position    INTEGER NOT NULL,
    start_ms    INTEGER NOT NULL DEFAULT 0,
    end_ms      INTEGER NOT NULL DEFAULT 0,
    text        TEXT NOT NULL
);

CREATE INDEX idx_lyrics_lines_document ON lyrics_lines(document_id, position);

CREATE TABLE lyrics_words (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    line_id       TEXT NOT NULL REFERENCES lyrics_lines(id) ON DELETE CASCADE,
    position      INTEGER NOT NULL,
    text          TEXT NOT NULL,
    start_ms      INTEGER NOT NULL DEFAULT 0,
    end_ms        INTEGER NOT NULL DEFAULT 0,
    timing_source TEXT NOT NULL DEFAULT 'provider'
);

CREATE INDEX idx_lyrics_words_line ON lyrics_words(line_id, position);

CREATE TABLE phonetic_annotations (
    id                TEXT PRIMARY KEY,
    line_id           TEXT NOT NULL REFERENCES lyrics_lines(id) ON DELETE CASCADE,
    accent            TEXT NOT NULL DEFAULT 'en-US',
    ipa               TEXT NOT NULL DEFAULT '',
    spanish_approx    TEXT NOT NULL DEFAULT '',
    translation       TEXT NOT NULL DEFAULT '',
    notes_json        TEXT NOT NULL DEFAULT '[]',
    model             TEXT NOT NULL DEFAULT '',
    generator_version TEXT NOT NULL DEFAULT '',
    cache_key         TEXT NOT NULL DEFAULT '',
    created_at        INTEGER NOT NULL
);

CREATE UNIQUE INDEX idx_phonetic_annotations_line ON phonetic_annotations(line_id, accent);

CREATE TABLE saved_vocabulary (
    id             TEXT PRIMARY KEY,
    word           TEXT NOT NULL,
    lemma          TEXT NOT NULL DEFAULT '',
    part_of_speech TEXT NOT NULL DEFAULT '',
    meaning        TEXT NOT NULL DEFAULT '',
    ipa            TEXT NOT NULL DEFAULT '',
    translation    TEXT NOT NULL DEFAULT '',
    example        TEXT NOT NULL DEFAULT '',
    source         TEXT NOT NULL DEFAULT 'ai',
    track_id       TEXT NOT NULL DEFAULT '',
    created_at     INTEGER NOT NULL
);

CREATE UNIQUE INDEX idx_saved_vocabulary_word ON saved_vocabulary(word);

CREATE TABLE learning_progress (
    track_id        TEXT PRIMARY KEY REFERENCES tracks(id) ON DELETE CASCADE,
    lines_practiced INTEGER NOT NULL DEFAULT 0,
    study_seconds   INTEGER NOT NULL DEFAULT 0,
    last_studied_at INTEGER NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'new'
);

CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE playback_history (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id              TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    played_at             INTEGER NOT NULL,
    duration_listened_ms  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_playback_history_time ON playback_history(played_at DESC);

CREATE TABLE provider_connections (
    provider         TEXT PRIMARY KEY,
    connection_state TEXT NOT NULL DEFAULT 'disconnected',
    metadata         TEXT NOT NULL DEFAULT '',
    updated_at       INTEGER NOT NULL
);
