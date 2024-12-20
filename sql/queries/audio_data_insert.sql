-- name: CreateGameArchive :exec
INSERT INTO game_archive (
    db_id, game_archive_id, tags, categories
) VALUES (
    ?, ?, ?, ?
);

-- name: CreateSoundbank :exec
INSERT INTO soundbank (
    db_id, toc_file_id, soundbank_path_name, soundbank_readable_name, categories, linked_game_archive_ids
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: CreateHierarchyObjectType :exec
INSERT INTO hierarchy_object_type (
    db_id, type
) VALUES (
    ?, ?
);

-- name: CreateHierarchyObject :exec
INSERT INTO hierarchy_object (
    db_id, wwise_object_id, type_db_id, parent_wwise_object_id, linked_soundbank_path_names
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: CreateRandomSeqContainer :exec
INSERT INTO random_seq_container (
    db_id, label, tags
) VALUES (
    ?, ?, ?
);

-- name: CreateSound :exec
INSERT INTO sound (
    db_id, wwise_short_id, label, tags
) VALUES (
    ?, ?, ?, ?
);

-- name: CreateWwiseStream :exec
INSERT INTO wwise_stream (
    db_id, toc_file_id, label, tags, linked_game_archive_ids
) VALUES (
    ?, ?, ?, ?, ?
);
