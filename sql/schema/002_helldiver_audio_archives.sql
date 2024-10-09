-- +goose Up
CREATE TABLE helldiver_audio_archives (
    id TEXT PRIMARY KEY,
    archive_id TEXT NOT NULL,
    basename TEXT NOT NULL,
    path TEXT NOT NULL,
    category TEXT
);

-- +goose Down
DROP TABLE helldiver_audio_archives;
