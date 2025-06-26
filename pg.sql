-- Polls table
CREATE TABLE polls (
    id UUID PRIMARY KEY,
    website_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
)

-- Poll options
CREATE TABLE poll_options (
    id UUID PRIMARY KEY,
    poll_id UUID REFERENCES polls(id),
    option_text TEXT NOT NULL
)

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    identifier TEXT UNIQUE NOT NULL -- email or phone
)

-- Votes table
CREATE TABLE votes (
    id UUID PRIMARY KEY,
    poll_id UUID REFERENCES polls(id),
    option_id UUID REFERENCES poll_options(id),
    user_id UUID REFERENCES users(id),
    voted_at TIMESTAMP NOT NULL DEFAULT NOW()
)