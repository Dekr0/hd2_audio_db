-- name: GetAllGameArchives :many
SELECT * FROM helldiver_game_archive;

-- name: GetAllHelldiverHirearchyObjectTypes :many
SELECT * FROM helldiver_hirearchy_object_type;

-- name: CreateHelldiverGameArchive :exec
INSERT INTO helldiver_game_archive (
    id, game_archive_id, tags, categories
) VALUES (
    ?, ?, ?, ?
);

-- name: CreateHelldiverSoundbank :exec
INSERT INTO helldiver_soundbank (
    id, toc_file_id, soundbank_path_name, soundbank_readable_name, categories, linked_game_archive_ids
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: CreateHelldiverHirearchyObjectType :exec
INSERT INTO helldiver_hirearchy_object_type (
    id, type
) VALUES (
    ?, ?
);

-- name: CreateHelldiverHirearchyObject :exec
INSERT INTO helldiver_hirearchy_object(
    id, wwise_object_id, type, parent_wwise_object_id, linked_soundbank_ids
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: CreateHelldiverWwiseStream :exec
INSERT INTO helldiver_wwise_stream (
    id, toc_file_id, label, tags, linked_game_archive_ids
) VALUES (
    ?, ?, ?, ?, ?
);

-- name: DeleteAllHelldiverGameArchive :exec
DELETE FROM helldiver_game_archive;

-- name: DeleteAllHelldiverSoundbank :exec
DELETE FROM helldiver_soundbank;

-- name: DeleteAllHelldiverHirearchyObjectType :exec
DELETE FROM helldiver_hirearchy_object_type;

-- name: DeleteAllHelldiverHirearchyObject :exec
DELETE FROM helldiver_hirearchy_object;

-- name: DeleteAllHelldiverWwiseStream :exec
DELETE FROM helldiver_wwise_stream;
