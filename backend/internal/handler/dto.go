package handler

import (
	"github.com/google/uuid"
	"time"
)

// Swagger DTO types for documentation

type errorResponse struct {
	Error   string `json:"error" example:"validation_error"`
	Message string `json:"message" example:"Email is required"`
}

type registerResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	CreatedAt time.Time `json:"created_at"`
}

type loginResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresAt   int64  `json:"expires_at" example:"1700000000"`
}

type documentResponse struct {
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID     uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Filename   string    `json:"filename" example:"report.pdf"`
	FileType   string    `json:"file_type" example:"pdf"`
	FileSize   int64     `json:"file_size" example:"102400"`
	ChunkCount int       `json:"chunk_count" example:"12"`
	CreatedAt  time.Time `json:"created_at"`
}

type documentListResponse struct {
	Documents []documentResponse `json:"documents"`
}

type sessionResponse struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Title     string    `json:"title" example:"What is the revenue?"`
	CreatedAt time.Time `json:"created_at"`
}

type sessionListResponse struct {
	Sessions []sessionResponse `json:"sessions"`
}

type messageResponse struct {
	ID           uuid.UUID   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SessionID    uuid.UUID   `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Role         string      `json:"role" example:"assistant"`
	Content      string      `json:"content" example:"According to the document..."`
	SourceChunks []uuid.UUID `json:"source_chunks,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

type messageListResponse struct {
	Messages []messageResponse `json:"messages"`
}

type healthResponse struct {
	Status string `json:"status" example:"ok"`
}
