CREATE VIEW IF NOT EXISTS sound_view 
AS 
SELECT 
hierarchy_object.wwise_object_id, 
hierarchy_object.parent_wwise_object_id, 
sound.wwise_short_id,
sound.label,
sound.tags,
hierarchy_object.linked_soundbank_path_names FROM
hierarchy_object INNER JOIN
sound ON hierarchy_object.db_id = sound.db_id;

CREATE VIEW IF NOT EXISTS random_seq_container_view
AS
SELECT
hierarchy_object.wwise_object_id,
hierarchy_object.parent_wwise_object_id,
random_seq_container.label,
random_seq_container.tags,
hierarchy_object.linked_soundbank_path_names FROM
hierarchy_object INNER JOIN
random_seq_container ON hierarchy_object.db_id = random_seq_container.db_id;
