package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Chunk struct {
	ID         uuid.UUID      `json:"id"`
	DocumentID uuid.UUID      `json:"document_id"`
	Content    string         `json:"content"`
	ChunkIndex int            `json:"chunk_index"`
	Embedding  pgvector.Vector `json:"-"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ChunkWithDocument struct {
	Chunk
	DocumentFilename string  `json:"document_filename"`
	Similarity       float64 `json:"similarity"`
}
