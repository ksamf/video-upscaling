CREATE TABLE IF NOT EXISTS languages (
    language_id SERIAL PRIMARY KEY,
    code VARCHAR(5) NOT NULL UNIQUE,
    language VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS videos (
    video_id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    language_id INT,
    quality INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_language FOREIGN KEY (language_id) REFERENCES languages (language_id)
);