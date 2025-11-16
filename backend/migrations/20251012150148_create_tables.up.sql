CREATE TABLE IF NOT EXISTS languages (
    language_id SERIAL PRIMARY KEY,
    code VARCHAR(5) NOT NULL UNIQUE,
    language VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    interests INT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS videos (
    video_id UUID PRIMARY KEY,
    user_id UUID REFERENCES users (user_id),
    name TEXT NOT NULL,
    language_id INT REFERENCES languages (language_id),
    quality INTEGER,
    views BIGINT DEFAULT 0,
    likes BIGINT DEFAULT 0,
    dislikes BIGINT DEFAULT 0,
    upscaled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS categories (
    category_id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS video_categories (
    video_id UUID REFERENCES videos (video_id),
    category_id INT REFERENCES categories (category_id),
    PRIMARY KEY (video_id, category_id)
);

CREATE TABLE IF NOT EXISTS user_video_actions (
    user_id UUID REFERENCES users (user_id),
    video_id UUID REFERENCES videos (video_id),
    liked BOOLEAN DEFAULT FALSE,
    disliked BOOLEAN DEFAULT FALSE,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, video_id)
);