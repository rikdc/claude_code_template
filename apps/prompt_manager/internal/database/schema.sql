-- Prompt Manager Database Schema
-- SQLite schema with PostgreSQL compatibility in mind

-- Conversations table - stores individual conversation threads
CREATE TABLE IF NOT EXISTS conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    title TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    prompt_count INTEGER DEFAULT 0,
    total_characters INTEGER DEFAULT 0,
    working_directory TEXT,
    transcript_path TEXT
);

-- Messages table - stores individual prompts and responses
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    message_type TEXT NOT NULL CHECK (message_type IN ('prompt', 'response')),
    content TEXT NOT NULL,
    character_count INTEGER DEFAULT 0,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tool_calls TEXT, -- JSON array of tool calls for responses
    execution_time INTEGER, -- milliseconds
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

-- Ratings table - stores user ratings for conversations or individual messages
CREATE TABLE IF NOT EXISTS ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER,
    message_id INTEGER,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    -- Ensure either conversation_id OR message_id is set, but not both
    CHECK ((conversation_id IS NOT NULL AND message_id IS NULL) OR 
           (conversation_id IS NULL AND message_id IS NOT NULL))
);

-- Tags table - stores available tags
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    color TEXT, -- hex color code for UI
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Conversation tags junction table - many-to-many relationship
CREATE TABLE IF NOT EXISTS conversation_tags (
    conversation_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (conversation_id, tag_id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Sessions table - tracks Claude Code sessions with metadata
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT UNIQUE NOT NULL,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    conversation_count INTEGER DEFAULT 0,
    total_prompt_count INTEGER DEFAULT 0,
    avg_response_time INTEGER DEFAULT 0,
    working_directory TEXT,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'completed', 'archived'))
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_conversations_session_id ON conversations(session_id);
CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_ratings_conversation_id ON ratings(conversation_id);
CREATE INDEX IF NOT EXISTS idx_ratings_message_id ON ratings(message_id);
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time);

-- Triggers to maintain conversation metadata
CREATE TRIGGER IF NOT EXISTS update_conversation_stats
    AFTER INSERT ON messages
    FOR EACH ROW
BEGIN
    UPDATE conversations 
    SET prompt_count = prompt_count + 1,
        total_characters = total_characters + NEW.character_count,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.conversation_id;
END;

CREATE TRIGGER IF NOT EXISTS update_conversation_timestamp
    AFTER UPDATE ON conversations
    FOR EACH ROW
BEGIN
    UPDATE conversations 
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;