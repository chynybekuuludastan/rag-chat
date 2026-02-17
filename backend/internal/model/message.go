package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatSession struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID           uuid.UUID   `json:"id"`
	SessionID    uuid.UUID   `json:"session_id"`
	Role         string      `json:"role"`
	Content      string      `json:"content"`
	SourceChunks []uuid.UUID `json:"source_chunks,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}
