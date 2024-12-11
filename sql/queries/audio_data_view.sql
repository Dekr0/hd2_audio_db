-- name: CreateSoundView :exec
CREATE VIEW sound_view
AS
SELECT hirearchy_object.wwise_object_id, wwise_short_id, parent_wwise_object_id, linked_soundbank_path_names
FROM sound
INNER JOIN hirearchy_object
ON sound.id = hirearchy_object.id;
