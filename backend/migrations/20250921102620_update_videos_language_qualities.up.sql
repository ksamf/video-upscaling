ALTER TABLE videos DROP COLUMN language;

ALTER TABLE videos DROP COLUMN qualities;

ALTER TABLE videos ADD COLUMN language_id INTEGER;

ALTER TABLE videos ADD COLUMN quality_id INTEGER;

ALTER TABLE videos
ADD CONSTRAINT fk_language FOREIGN KEY (language_id) REFERENCES languages (language_id);

ALTER TABLE videos
ADD CONSTRAINT fk_quality FOREIGN KEY (quality_id) REFERENCES qualities (quality_id);