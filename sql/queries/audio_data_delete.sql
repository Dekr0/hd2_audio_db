-- name: DeleteAllGameArchive :exec
DELETE FROM game_archive;

-- name: DeleteAllSoundbank :exec
DELETE FROM soundbank;

-- name: DeleteAllHierarchyObjectType :exec
DELETE FROM hierarchy_object_type;

-- name: DeleteAllHierarchyObject :exec
DELETE FROM hierarchy_object;

-- name: DeleteAllWwiseStream :exec
DELETE FROM wwise_stream;

-- name: DeleteAllSound :exec
DELETE FROM sound;

-- name: DeleteAllRandomSeqContainer :exec
DELETE FROM random_seq_container;
