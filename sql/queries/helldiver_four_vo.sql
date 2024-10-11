-- name: CreateHelldiverFourTranscription :exec
INSERT INTO helldiver_four_voice_line (
    id, file_id, transcription
) VALUES (
    ?, ?, ?
);

-- name: CreateHelldiverFourVOFTSEntry :exec
INSERT INTO helldiver_four_vo_fts (
    id, file_id, transcription
) VALUES (
    ?, ?, ?
);
