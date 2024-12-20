-- +goose Up
CREATE TABLE game_archive (
    db_id TEXT PRIMARY KEY,
    game_archive_id TEXT NOT NULL UNIQUE,
    tags TEXT NOT NULL,
    categories TEXT NOT NULL
);

CREATE TABLE soundbank (
    db_id TEXT PRIMARY KEY,
    toc_file_id TEXT NOT NULL UNIQUE,
    soundbank_path_name TEXT NOT NULL,
    soundbank_readable_name TEXT NOT NULL,
    categories TEXT NOT NULL,
    linked_game_archive_ids TEXT NOT NULL
);

CREATE TABLE hierarchy_object_type (
    db_id TEXT PRIMARY KEY,
    type TEXT NOT NULL UNIQUE
);

CREATE TABLE hierarchy_object (
    db_id TEXT PRIMARY KEY,
    wwise_object_id TEXT NOT NULL UNIQUE,
    type_db_id TEXT NOT NULL,
    parent_wwise_object_id TEXT NOT NULL,
    linked_soundbank_path_names TEXT NOT NULL,
    FOREIGN KEY (type_db_id) REFERENCES hierarchy_object_type(db_id)
);

CREATE TABLE random_seq_container (
    db_id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    FOREIGN KEY (db_id) REFERENCES hierarchy_object(db_id)
);

CREATE TABLE sound (
    db_id TEXT NOT NULL,
    wwise_short_id TEXT NOT NULL,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    PRIMARY KEY (db_id, wwise_short_id),
    FOREIGN KEY (db_id) REFERENCES hierarchy_object(db_id)
);

CREATE TABLE wwise_stream (
    db_id TEXT PRIMARY KEY,
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
DROP TABLE hierarchy_object;
DROP TABLE hierarchy_object_type;
DROP TABLE game_archive;
