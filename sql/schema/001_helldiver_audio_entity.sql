-- +goose Up
CREATE TABLE archive (
    aid TEXT PRIMARY KEY,
    tags TEXT NOT NULL,
    categories TEXT NOT NULL,
    date_modified TEXT NOT NULL
);

CREATE TABLE asset (
    aid TEXT NOT NULL,
    fid INTEGER NOT NULL,
    tid INTEGER NOT NULL,
    data_offset INTEGER NOT NULL,
    stream_file_offset INTEGER NOT NULL,
    gpu_rsrc_offset INTEGER NOT NULL,
    unknown_01 INTEGER NOT NULL,
    unknown_02 INTEGER NOT NULL,
    data_size INTEGER NOT NULL,
    stream_size INTEGER NOT NULL,
    gpu_rsrc_size INTEGER NOT NULL,
    unknown_03 INTEGER NOT NULL,
    unknown_04 INTEGER NOT NULL,
    PRIMARY KEY (aid, fid, tid),
    FOREIGN KEY (aid) REFERENCES archive(aid)
);

CREATE TABLE soundbank (
    aid TEXT NOT NULL,
    fid INTEGER NOT NULL,
    path TEXT NOT NULL,
    name TEXT NOT NULL,
    categories TEXT NOT NULL,
    PRIMARY KEY (aid, fid),
    FOREIGN KEY (aid) REFERENCES archive(aid),
    FOREIGN KEY (fid) REFERENCES asset(fid)
);

CREATE TABLE hierarchy (
    aid TEXT NOT NULL,
    fid INTEGER NOT NULL,
    hid INTEGER NOT NULL,
    type TEXT NOT NULL,
    parent INTEGER NOT NULL,
    label TEXT NOT NULL,
    tags TEXT NOT NULL,
    description TEXT NOT NULL,
    PRIMARY KEY (aid, fid, hid, type),
    FOREIGN KEY (aid) REFERENCES archive(aid),
    FOREIGN KEY (fid) REFERENCES asset(fid)
);

CREATE TABLE sound (
    aid TEXT NOT NULL,
    fid INTEGER NOT NULL,
    hid INTEGER NOT NULL,
    sid INTEGER NOT NULL,
    PRIMARY KEY (aid, fid, hid),
    FOREIGN KEY (aid) REFERENCES archive(aid),
    FOREIGN KEY (fid) REFERENCES asset(fid),
    FOREIGN KEY (hid) REFERENCES hierarchy(hid)
);

-- +goose Down
DROP TABLE sound;
DROP TABLE hierarchy;
DROP TABLE soundbank;
DROP TABLE asset;
DROP TABLE archive;
