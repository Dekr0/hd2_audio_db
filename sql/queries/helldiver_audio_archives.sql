-- name: CreateHelldiverAudioArchive :exec
INSERT INTO helldiver_audio_archives (
    id, archive_id, basename, path, category
) VALUES (
    ?, ?, ?, ?, ?
);
