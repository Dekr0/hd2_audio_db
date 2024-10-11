-- name: CreateHelldiverAudioArchive :exec
INSERT INTO helldiver_audio_archive (
    id, archive_id, tag, category
) VALUES (
    ?, ?, ?, ? 
);
