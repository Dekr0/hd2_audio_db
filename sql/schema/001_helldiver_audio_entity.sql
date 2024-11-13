-- +goose Up
CREATE TABLE helldiver_game_archive (
    id TEXT PRIMARY KEY,
    game_archive_id TEXT NOT NULL UNIQUE,
    tags TEXT NOT NULL,
    categories TEXT NOT NULL
);

CREATE TABLE helldiver_soundbank (
    id TEXT PRIMARY KEY,
    soundbank_id TEXT NOT NULL UNIQUE,
    soundbank_name TEXT NOT NULL UNIQUE,
    soundbank_readable_name TEXT NOT NULL,
    categories TEXT NOT NULL,
    linked_game_archive_ids TEXT NOT NULL
);

CREATE TABLE helldiver_audio_source (
    id TEXT PRIMARY KEY,
    audio_source_id TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    linked_soundbank_ids TEXT NOT NULL
);

-- +goose Down
DROP TABLE helldiver_audio_source;
DROP TABLE helldiver_soundbank;
DROP TABLE helldiver_game_archive;
