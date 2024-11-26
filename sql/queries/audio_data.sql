-- name: GetAllGameArchives :many
SELECT * FROM helldiver_game_archive;

-- name: CreateHelldiverGameArchive :exec
INSERT INTO helldiver_game_archive (
    id, game_archive_id, tags, categories
) VALUES (
    ?, ?, ?, ?
);

-- name: CreateHelldiverSoundbank :exec
INSERT INTO helldiver_soundbank (
    id, toc_file_id, sonndbank_path_name, soundbank_readable_name, categories, linked_game_archive_ids
) VALUES (
    ?, ?, ?, ?, ?, ?
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

-- name: DeleteAllHelldiverWwiseStream :exec
DELETE FROM helldiver_wwise_stream
