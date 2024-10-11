-- +goose Up
CREATE TABLE helldiver_four_voice_line (
    id TEXT PRIMARY KEY,
    file_id TEXT NOT NULL,
    transcription TEXT NOT NULL
);

CREATE VIRTUAL TABLE helldiver_four_vo_fts USING fts5(id, file_id, transcription);

-- +goose Down
DROP TABLE helldiver_four_voice_line;

DROP TABLE helldiver_four_vo_fts;
