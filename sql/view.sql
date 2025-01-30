CREATE VIEW IF NOT EXISTS sound_view AS
SELECT
    hierarchy_object.wwise_object_id,
    R2.wwise_short_id,
    hierarchy_object.label,
    hierarchy_object.tags,
    hierarchy_object.description,
    R2.toc_file_id,
    R2.soundbank_path_name
    FROM hierarchy_object
INNER JOIN (
    SELECT 
        sound.db_id,
        sound.wwise_short_id,
        R1.toc_file_id,
        R1.soundbank_path_name
    FROM sound 
    INNER JOIN (
            SELECT 
                hierarchy_object_db_id, 
                soundbank.toc_file_id,
                soundbank.soundbank_path_name
            FROM soundbank_hierarchy_object_relation
            INNER JOIN soundbank 
            ON soundbank_hierarchy_object_relation.soundbank_db_id = soundbank.db_id
        ) AS R1 
        ON sound.db_id = R1.hierarchy_object_db_id
) AS R2 
ON hierarchy_object.db_id = R2.db_id;


CREATE VIEW IF NOT EXISTS hierarchy_object_view AS
SELECT
    R2.wwise_object_id,
    hierarchy_object_type.type,
    R2.parent_wwise_object_id,
    R2.label,
    R2.tags,
    R2.description,
    R2.toc_file_id,
    R2.soundbank_path_name
FROM hierarchy_object_type
INNER JOIN
(
    SELECT
        hierarchy_object.wwise_object_id,
        hierarchy_object.type_db_id,
        hierarchy_object.parent_wwise_object_id,
        hierarchy_object.label,
        hierarchy_object.tags,
        hierarchy_object.description,
        R.toc_file_id,
        R.soundbank_path_name
    FROM hierarchy_object
    INNER JOIN (
        SELECT 
            hierarchy_object_db_id, 
            soundbank.toc_file_id,
            soundbank.soundbank_path_name 
        FROM soundbank_hierarchy_object_relation
        INNER JOIN soundbank 
        ON soundbank_hierarchy_object_relation.soundbank_db_id = soundbank.db_id
    ) AS R 
    ON hierarchy_object.db_id = R.hierarchy_object_db_id
) AS R2
ON hierarchy_object_type.db_id = R2.type_db_id;

CREATE VIEW IF NOT EXISTS soundbank_view AS
SELECT
    R.game_archive_id,
    R.tags AS game_archive_tags,
    R.categories AS game_archive_categories,
    toc_file_id,
    soundbank_path_name,
    soundbank_readable_name
FROM soundbank INNER JOIN
(
    SELECT soundbank_db_id, game_archive_id, tags, categories
    FROM game_archive_soundbank_relation
    INNER JOIN game_archive
    ON game_archive.db_id = game_archive_db_id
) AS R
ON R.soundbank_db_id = soundbank.db_id;
