CREATE TABLE videos (
    id        SERIAL      PRIMARY KEY,
    name      TEXT        NOT NULL,
    url       TEXT        NOT NULL,
    language  TEXT,
    nsfw      BOOLEAN     DEFAULT FALSE,
    qualities INT[]       
);
