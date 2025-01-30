-- name: UpdateLabelByWwiseObjectId :exec
UPDATE hierarchy_object SET label = ? WHERE wwise_object_id = ?;

-- name: UpdateDescriptionByWwiseObjectId :exec
UPDATE hierarchy_object SET description = ? WHERE wwise_object_id = ?;

-- name: UpdateInfoByWwiseObjectId :exec
UPDATE hierarchy_object SET label = ?, description = ? WHERE wwise_object_id = ?;

-- name: UpdateLabelByWwiseShortId :exec
UPDATE hierarchy_object SET label = ?
WHERE hierarchy_object.db_id IN (
    SELECT sound.db_id FROM sound 
    WHERE sound.wwise_short_id = ?
);

-- name: UpdateDescByWwiseShortId :exec
UPDATE hierarchy_object SET description = ?
WHERE hierarchy_object.db_id IN (
    SELECT sound.db_id FROM sound
    WHERE sound.wwise_short_id = ?
);

-- name: UpdateInfoByWwiseShortId :exec
UPDATE hierarchy_object SET label = ?, description = ?
WHERE hierarchy_object.db_id IN (
    SELECT sound.db_id FROM sound
    WHERE sound.wwise_short_id = ?
);
