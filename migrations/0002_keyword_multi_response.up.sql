CREATE TABLE keyword_responses (
    id SERIAL PRIMARY KEY,
    keyword_id INTEGER NOT NULL REFERENCES keywords (id) ON DELETE CASCADE,
    response TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (keyword_id, response)
);

INSERT INTO keyword_responses (keyword_id, response)
SELECT id, response FROM keywords;

ALTER TABLE keywords DROP COLUMN response;
