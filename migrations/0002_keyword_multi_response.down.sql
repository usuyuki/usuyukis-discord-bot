ALTER TABLE keywords ADD COLUMN response TEXT;

UPDATE keywords
SET response = kr.response
FROM (
    SELECT DISTINCT ON (keyword_id) keyword_id, response
    FROM keyword_responses
    ORDER BY keyword_id, id
) kr
WHERE keywords.id = kr.keyword_id;

ALTER TABLE keywords ALTER COLUMN response SET NOT NULL;

DROP TABLE keyword_responses;
