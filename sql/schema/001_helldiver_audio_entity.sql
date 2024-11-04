-- +goose Up
CREATE TABLE helldiver_game_archive (
    id TEXT PRIMARY KEY,
    game_archive_id TEXT NOT NULL UNIQUE,
    categories TEXT NOT NULL
);

CREATE TABLE helldiver_audio_bank (
    id TEXT PRIMARY KEY,
    audio_bank_id TEXT NOT NULL UNIQUE,
    audio_bank_name TEXT NOT NULL,
    category TEXT NOT NULL,
    linked_game_archive_ids TEXT NOT NULL
);

CREATE TABLE helldiver_audio_source (
    id TEXT PRIMARY KEY,
    audio_source_id TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    linked_audio_bank_id TEXT NOT NULL
);

-- +goose Down
DROP TABLE helldiver_audio_source;
DROP TABLE helldiver_audio_bank;
DROP TABLE helldiver_game_archive;
