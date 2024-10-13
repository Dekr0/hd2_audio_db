-- +goose Up
CREATE TABLE helldiver_audio_archive_name (
    audio_archive_name_id TEXT PRIMARY KEY,
    audio_archive_name TEXT NOT NULL
);

CREATE TABLE helldiver_audio_archive (
    audio_archive_db_id TEXT PRIMARY KEY,
    audio_archive_id TEXT NOT NULL,
    audio_archive_name_id TEXT NOT NULL,
    audio_archive_category TEXT NOT NULL,
    FOREIGN KEY(audio_archive_name_id) REFERENCES helldiver_audio_archive_name(audio_archive_name_id)
);

-- +goose Down
DROP TABLE helldiver_audio_archive_name;

DROP TABLE helldiver_audio_archive;
