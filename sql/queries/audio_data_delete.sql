-- name: DeleteAllGameArchive :exec
DELETE FROM game_archive;

-- name: DeleteAllSoundbank :exec
DELETE FROM soundbank;

-- name: DeleteAllArchiveSoundbankRelation :exec
DELETE FROM game_archive_soundbank_relation;

-- name: DeleteAllHierarchyObjectType :exec
DELETE FROM hierarchy_object_type;

-- name: DeleteAllHierarchyObject :exec
DELETE FROM hierarchy_object;

-- name: DeletionAllHierarchyObjectRelation :exec
DELETE FROM soundbank_hierarchy_object_relation;

-- name: DeleteAllWwiseStream :exec
DELETE FROM wwise_stream;

-- name: DeleteAllSound :exec
DELETE FROM sound;
