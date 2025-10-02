CREATE TABLE IF NOT EXISTS videos (
    video_id UUID,
    name TEXT NOT NULL,
    video_path TEXT NOT NULL,
    language TEXT,
    qualities INT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);