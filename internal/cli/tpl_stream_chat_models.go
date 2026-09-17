package cli

const tplConversationModel = `package models

import "time"

type Conversation struct {
	ID        int64
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
`

const tplMessageModel = `package models

import "time"

type Message struct {
	ID             int64
	ConversationID int64
	Role           string
	Content        string
	CreatedAt      time.Time
}
`

const tplChatMigration = `-- up
CREATE TABLE conversations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);

-- down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
`
