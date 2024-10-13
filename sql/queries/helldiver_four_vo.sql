-- name: CreateHelldiverFourTranscription :exec
INSERT INTO helldiver_four_voice_line (
    helldiver_four_voice_line_id, file_id, transcription
) VALUES (
    ?, ?, ?
);

-- name: CreateHelldiverFourVOFTSEntry :exec
INSERT INTO helldiver_four_vo_fts (
    helldiver_four_voice_line_id, file_id, transcription
) VALUES (
    ?, ?, ?
);
