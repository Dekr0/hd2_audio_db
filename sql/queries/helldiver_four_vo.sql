-- name: CreateHelldiverFourTranscription :exec
INSERT INTO helldiver_four_voice_lines (
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
