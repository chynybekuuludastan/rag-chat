package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Document, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Document, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateChunkCount(ctx context.Context, id uuid.UUID, count int) error
}

type ChunkRepository interface {
	CreateBatch(ctx context.Context, chunks []model.Chunk) error
	SearchSimilar(ctx context.Context, embedding []float32, userID uuid.UUID, limit int, threshold float64) ([]model.ChunkWithDocument, error)
	DeleteByDocument(ctx context.Context, documentID uuid.UUID) error
}

type ChatRepository interface {
	CreateSession(ctx context.Context, session *model.ChatSession) error
	GetSession(ctx context.Context, id uuid.UUID) (*model.ChatSession, error)
	ListSessionsByUser(ctx context.Context, userID uuid.UUID) ([]model.ChatSession, error)
	UpdateSessionTitle(ctx context.Context, id uuid.UUID, title string) error
	CreateMessage(ctx context.Context, msg *model.Message) error
	GetMessagesBySession(ctx context.Context, sessionID uuid.UUID) ([]model.Message, error)
}
