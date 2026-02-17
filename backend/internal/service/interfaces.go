package service

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type ChatEvent struct {
	Type    string      `json:"type"`
	Content string      `json:"content,omitempty"`
	Sources []SourceRef `json:"chunks,omitempty"`
}

type SourceRef struct {
	ID               uuid.UUID `json:"id"`
	DocumentFilename string    `json:"document"`
	Content          string    `json:"content"`
	Similarity       float64   `json:"similarity"`
}

type AuthService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
}

type DocumentService interface {
	Upload(ctx context.Context, userID uuid.UUID, file io.Reader, filename string, fileSize int64) (*model.Document, error)
	List(ctx context.Context, userID uuid.UUID) ([]model.Document, error)
	Delete(ctx context.Context, userID uuid.UUID, docID uuid.UUID) error
}

type MessageWithSources struct {
	model.Message
	Sources []SourceRef `json:"sources,omitempty"`
}

type ChatService interface {
	Ask(ctx context.Context, userID uuid.UUID, sessionID *uuid.UUID, question string) (*uuid.UUID, <-chan ChatEvent, error)
	GetHistory(ctx context.Context, userID uuid.UUID) ([]model.ChatSession, error)
	GetMessages(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) ([]MessageWithSources, error)
}
