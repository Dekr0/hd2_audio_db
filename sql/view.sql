CREATE VIEW IF NOT EXISTS hierarchy_view AS
SELECT
    hierarchy.aid,
    hierarchy.fid,
    soundbank.path,
    hierarchy.hid,
    hierarchy.type,
    hierarchy.parent,
    hierarchy.label,
    hierarchy.tags
FROM hierarchy
INNER JOIN soundbank
ON hierarchy.aid = soundbank.aid AND hierarchy.fid = soundbank.fid;
