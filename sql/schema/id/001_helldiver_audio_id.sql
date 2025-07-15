-- +goose Up
CREATE TABLE hierarchy (
    hid INTEGER NOT NULL,
    PRIMARY KEY (hid)
);

CREATE TABLE source (
    sid INTEGER NOT NULL,
    PRIMARY KEY (sid)
);

-- +goose Down
DROP TABLE hierarchy;
DROP TABLE source;
