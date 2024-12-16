CREATE VIEW sound_view 
AS 
SELECT 
hirearchy_object.wwise_object_id, 
hirearchy_object.parent_wwise_object_id, 
sound.wwise_short_id,
sound.label,
sound.tags,
hirearchy_object.linked_soundbank_path_names FROM
hirearchy_object INNER JOIN
sound ON hirearchy_object.id = sound.id;
