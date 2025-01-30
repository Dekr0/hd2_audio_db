-- name: CreateGameArchive :exec
INSERT INTO game_archive (
    db_id, game_archive_id, tags, categories
) VALUES (
    ?, ?, ?, ?
);

-- name: CreateSoundbank :exec
INSERT INTO soundbank (
    db_id, toc_file_id, soundbank_path_name, soundbank_readable_name, categories
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: CreateArchiveSounbankRelation :exec
INSERT INTO game_archive_soundbank_relation (
    game_archive_db_id, soundbank_db_id
) VALUES (
    ?, ?
);

-- name: CreateHierarchyObjectType :exec
INSERT INTO hierarchy_object_type (
    db_id, type
) VALUES (
    ?, ?
);

-- name: CreateHierarchyObject :exec
INSERT INTO hierarchy_object (
    db_id, wwise_object_id, type_db_id, parent_wwise_object_id, label, tags, description
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
);

-- name: CreateSoundbankHierarchyObjectRelation :exec
INSERT INTO soundbank_hierarchy_object_relation (
    soundbank_db_id, hierarchy_object_db_id
) VALUES (
    ?, ?
);

-- name: CreateSound :exec
INSERT INTO sound (
    db_id, wwise_short_id
) VALUES (
    ?, ?
);
