-- name: InsertArchive :exec
INSERT INTO archive (
    aid, tags, categories, date_modified
) VALUES (
    ?, ?, ?, ?
);

-- name: InsertAsset :exec
INSERT INTO asset (
    aid, fid, tid, 
    data_offset, stream_file_offset, gpu_rsrc_offset,
    unknown_01, unknown_02,
    data_size, stream_size, gpu_rsrc_size,
    unknown_03, unknown_04
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertSoundbank :exec
INSERT INTO soundbank (aid, fid, path, name, categories) VALUES (?, ?, ?, ?, ?);

-- name: InsertHierarchy :exec
INSERT INTO hierarchy (
    aid, fid, hid,
    type, parent,
    label, tags, description
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertSound :exec
INSERT INTO sound (aid, fid, hid, sid) VALUES (?, ?, ?, ?);
