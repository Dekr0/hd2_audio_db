-- name: HierarchyId :one
SELECT COUNT(*) FROM hierarchy WHERE hid = ?; 

-- name: SourceId :one
SELECT COUNT(*) FROM source WHERE sid = ?;
