-- name: GetAllArchive :many
SELECT * FROM archive;

-- name: GetAllSoundbank :many
SELECT * FROM soundbank;

-- name: HierarchyIdUnique :many
SELECT hid FROM hierarchy GROUP BY hid HAVING COUNT(*) = 1;

-- name: SourceIdUnique :many
SELECT sid FROM sound GROUP BY sid HAVING COUNT(*) = 1;
