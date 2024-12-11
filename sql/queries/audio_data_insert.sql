-- name: CreateGameArchive :exec
INSERT INTO game_archive (
    id, game_archive_id, tags, categories
) VALUES (
    ?, ?, ?, ?
);

-- name: CreateSoundbank :exec
INSERT INTO soundbank (
    id, toc_file_id, soundbank_path_name, soundbank_readable_name, categories, linked_game_archive_ids
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: CreateHirearchyObjectType :exec
INSERT INTO hirearchy_object_type (
    id, type
) VALUES (
    ?, ?
);

-- name: CreateHirearchyObject :exec
INSERT INTO hirearchy_object (
    id, wwise_object_id, type, parent_wwise_object_id, linked_soundbank_path_names
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: CreateRandomSeqContainer :exec
INSERT INTO random_seq_container (
    id, label, tags
) VALUES (
    ?, ?, ?
);

-- name: CreateSound :exec
INSERT INTO sound (
    id, wwise_short_id, label, tags
) VALUES (
    ?, ?, ?, ?
);

-- name: CreateWwiseStream :exec
INSERT INTO wwise_stream (
    id, toc_file_id, label, tags, linked_game_archive_ids
) VALUES (
    ?, ?, ?, ?, ?
);
