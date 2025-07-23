-- Initial schema migration
-- Version: 001
-- Description: Create core tables for conversations, messages, ratings, tags, and sessions

-- Conversations table
CREATE TABLE conversations (
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

-- Messages table
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    message_type TEXT NOT NULL CHECK (message_type IN ('prompt', 'response')),
    content TEXT NOT NULL,
    character_count INTEGER DEFAULT 0,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    tool_calls TEXT,
    execution_time INTEGER,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

-- Ratings table
CREATE TABLE ratings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER,
    message_id INTEGER,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    CHECK ((conversation_id IS NOT NULL AND message_id IS NULL) OR 
           (conversation_id IS NULL AND message_id IS NOT NULL))
);

-- Tags table
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    color TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Conversation tags junction table
CREATE TABLE conversation_tags (
    conversation_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (conversation_id, tag_id),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Sessions table
CREATE TABLE sessions (
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

-- Indexes
CREATE INDEX idx_conversations_session_id ON conversations(session_id);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
CREATE INDEX idx_ratings_conversation_id ON ratings(conversation_id);
CREATE INDEX idx_ratings_message_id ON ratings(message_id);
CREATE INDEX idx_sessions_session_id ON sessions(session_id);
CREATE INDEX idx_sessions_start_time ON sessions(start_time);

-- Triggers
CREATE TRIGGER update_conversation_stats
    AFTER INSERT ON messages
    FOR EACH ROW
BEGIN
    UPDATE conversations 
    SET prompt_count = prompt_count + 1,
        total_characters = total_characters + NEW.character_count,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.conversation_id;
END;

CREATE TRIGGER update_conversation_timestamp
    AFTER UPDATE ON conversations
    FOR EACH ROW
BEGIN
    UPDATE conversations 
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;