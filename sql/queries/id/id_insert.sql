-- name: InsertHierarchy :exec
INSERT INTO hierarchy (hid) VALUES (?);

-- name: InsertSource :exec
INSERT INTO source (sid) VALUES (?);
