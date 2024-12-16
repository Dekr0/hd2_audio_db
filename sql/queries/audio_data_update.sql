-- name: UpdateSoundLabelBySourceId :exec
UPDATE sound SET label = ? WHERE wwise_short_id = ?;
