-- 0002_player_only.sql
-- The app is a player with phonetic transliteration and translation, not a
-- study tool: the learning tables and the generated demo catalog go away.
-- Demo rows are deleted first so their lyric documents, playlists entries and
-- history entries leave with them.

DELETE FROM playback_history
WHERE track_id IN (SELECT id FROM tracks WHERE provider = 'demo');

DELETE FROM playlist_items
WHERE track_id IN (SELECT id FROM tracks WHERE provider = 'demo');

DELETE FROM library_tracks
WHERE track_id IN (SELECT id FROM tracks WHERE provider = 'demo');

DELETE FROM phonetic_annotations
WHERE line_id IN (
    SELECT lyrics_lines.id
    FROM lyrics_lines
    JOIN lyrics_documents ON lyrics_documents.id = lyrics_lines.document_id
    WHERE lyrics_documents.track_id IN (SELECT id FROM tracks WHERE provider = 'demo')
);

DELETE FROM lyrics_documents
WHERE track_id IN (SELECT id FROM tracks WHERE provider = 'demo');

DELETE FROM tracks WHERE provider = 'demo';

DROP TABLE IF EXISTS saved_vocabulary;
DROP TABLE IF EXISTS learning_progress;
