-- name: DeleteAllArchive :exec
DELETE FROM archive;

-- name: DeleteAllAsset :exec
DELETE FROM asset;

-- name: DeleteAllSoundbank :exec
DELETE FROM soundbank;

-- name: DeleteAllHierarchy :exec
DELETE FROM hierarchy;

-- name: DeleteAllSound :exec
DELETE FROM sound;
