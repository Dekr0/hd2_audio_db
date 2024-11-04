-- name: CreateHelldiverGameArchive :exec
INSERT INTO helldiver_game_archive (
    id, game_archive_id, categories
) VALUES (
    ?, ?, ?
);

-- name: DeleteAllHelldiverGameArchive :exec
DELETE FROM helldiver_game_archive;
