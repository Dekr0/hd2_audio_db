-- name: DeleteAllGameArchive :exec
DELETE FROM game_archive;

-- name: DeleteAllSoundbank :exec
DELETE FROM soundbank;

-- name: DeleteAllHirearchyObjectType :exec
DELETE FROM hirearchy_object_type;

-- name: DeleteAllHirearchyObject :exec
DELETE FROM hirearchy_object;

-- name: DeleteAllWwiseStream :exec
DELETE FROM wwise_stream;

-- name: DeleteAllSound :exec
DELETE FROM sound;

-- name: DeleteAllRandomSeqContainer :exec
DELETE FROM random_seq_container;
