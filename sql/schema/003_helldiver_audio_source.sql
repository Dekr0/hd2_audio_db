-- +goose Up
CREATE TABLE helldiver_audio_source (
    id TEXT PRIMARY KEY,
    audio_source_id TEXT NOT NULL,
    relative_archive_ids TEXT NOT NULL,
    relative_archive_tags TEXT NOT NULL
);

-- +goose Down
DROP TABLE helldiver_audio_source;
