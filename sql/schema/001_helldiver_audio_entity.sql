-- +goose Up
CREATE TABLE game_archive (
    id TEXT PRIMARY KEY,
    game_archive_id TEXT NOT NULL UNIQUE,
    tags TEXT NOT NULL,
    categories TEXT NOT NULL
);

CREATE TABLE soundbank (
    id TEXT PRIMARY KEY,
    toc_file_id TEXT NOT NULL UNIQUE,
    soundbank_path_name TEXT NOT NULL UNIQUE,
    soundbank_readable_name TEXT NOT NULL,
    categories TEXT NOT NULL,
    linked_game_archive_ids TEXT NOT NULL
);

CREATE TABLE hirearchy_object_type (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL UNIQUE
);

CREATE TABLE hirearchy_object (
    id TEXT PRIMARY KEY,
    wwise_object_id TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    parent_wwise_object_id TEXT NOT NULL,
    linked_soundbank_path_names TEXT NOT NULL,
    FOREIGN KEY (type) REFERENCES hirearchy_object_type(id)
);

CREATE TABLE random_seq_container (
    id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    FOREIGN KEY (id) REFERENCES hirearchy_object(id)
);

CREATE TABLE sound (
    id TEXT NOT NULL,
    wwise_short_id TEXT NOT NULL,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    PRIMARY KEY (id, wwise_short_id),
    FOREIGN KEY (id) REFERENCES hirearchy_object(id)
);

CREATE TABLE wwise_stream (
    id PRIMARY KEY,
    toc_file_id TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    linked_game_archive_ids TEXT NOT NULL
);

-- +goose Down
DROP TABLE wwise_stream;
DROP TABLE sound;
DROP TABLE random_seq_container;
DROP TABLE soundbank;
DROP TABLE hirearchy_object;
DROP TABLE hirearchy_object_type;
DROP TABLE game_archive;
