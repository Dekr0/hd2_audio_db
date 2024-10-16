-- name: CreateHelldiverAudioArchive :exec
INSERT INTO helldiver_audio_archive (
    audio_archive_db_id, audio_archive_id, audio_archive_name_id, audio_archive_category
) VALUES (
    ?, ?, ?, ? 
);

-- name: CreateHelldiverAudioArchiveName :exec
INSERT INTO helldiver_audio_archive_name (
    audio_archive_name_id, audio_archive_name
    ) VALUES (
    ?, ?
);

-- name: DeleteAllHelldiverAudioArchive :exec
DELETE FROM helldiver_audio_archive;

-- name: DeleteAllHelldiverArchiveName :exec
DELETE FROM helldiver_audio_archive_name;

-- name: DeleteAllHelldiverAudioSource :exec
DELETE FROM helldiver_audio_source;

-- name: HasAudioArchiveID :one
SELECT DISTINCT COUNT(audio_archive_id) FROM helldiver_audio_archive 
WHERE audio_archive_id = ?;

-- name: QueryAllAudioArchiveName :many
SELECT DISTINCT audio_archive_name_id, audio_archive_name FROM 
helldiver_audio_archive_name;

-- name: QueryAudioArchiveNameByCategory :many
SELECT DISTINCT helldiver_audio_archive.audio_archive_name_id, audio_archive_name
FROM helldiver_audio_archive INNER JOIN helldiver_audio_archive_name ON 
helldiver_audio_archive.audio_archive_name_id = helldiver_audio_archive_name.audio_archive_name_id
WHERE audio_archive_category = ?;

-- name: QuerySharedAudioSourceByAudioSourceID :many
SELECT audio_source_id, linked_audio_archive_ids, linked_audio_archive_name_ids FROM 
helldiver_audio_source 
WHERE audio_source_id IN (sqlc.slice('audio_source_ids')) AND
linked_audio_archive_ids LIKE '%,%';

-- name: QueryAllSharedAudioSourceByAudioArchiveID :many
SELECT audio_source_id, linked_audio_archive_ids, linked_audio_archive_name_ids 
FROM helldiver_audio_source
WHERE linked_audio_archive_ids LIKE '%' || ? || ',%';

-- name: QueryAllSafeAudioSourceByAudioArchiveID :many
SELECT audio_source_id, linked_audio_archive_ids, linked_audio_archive_name_ids 
FROM helldiver_audio_source
WHERE linked_audio_archive_ids = ?;

-- name: QueryAllSharedAudioSourceByAudioArchiveNameID :many
SELECT audio_source_id, linked_audio_archive_ids, linked_audio_archive_name_ids 
FROM helldiver_audio_source 
WHERE linked_audio_archive_name_ids LIKE '%' || ? || ',%';

-- name: QueryAllSafeAudioSourceByAudioArchiveNameID :many
SELECT audio_source_id, linked_audio_archive_ids, linked_audio_archive_name_ids 
FROM helldiver_audio_source 
WHERE linked_audio_archive_name_ids = ?;
