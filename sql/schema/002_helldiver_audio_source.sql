-- +goose Up
CREATE TABLE helldiver_audio_source (
    audio_source_db_id TEXT PRIMARY KEY,
    audio_source_id TEXT NOT NULL,
    linked_audio_archive_ids TEXT NOT NULL,
    linked_audio_archive_name_ids TEXT NOT NULL
);

-- +goose Down
DROP TABLE helldiver_audio_source;
