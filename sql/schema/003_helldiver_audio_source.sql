-- +goose Up
CREATE TABLE helldiver_audio_source (
    id TEXT PRIMARY KEY,
    audio_source_id INTEGER NOT NULL,
    relative_archive_ids BLOB NOT NULL
);

-- +goose Down
DROP TABLE helldiver_audio_source;
