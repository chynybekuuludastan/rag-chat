package service

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
	"github.com/dastanchynybek/rag-chat/backend/internal/pkg/chunker"
	"github.com/dastanchynybek/rag-chat/backend/internal/pkg/llm"
	"github.com/dastanchynybek/rag-chat/backend/internal/pkg/parser"
	"github.com/dastanchynybek/rag-chat/backend/internal/repository"
)

const maxFileSize = 10 * 1024 * 1024 // 10MB

type documentService struct {
	docRepo   repository.DocumentRepository
	chunkRepo repository.ChunkRepository
	llm       llm.LLMClient
	chunkerCfg chunker.Config
}

func NewDocumentService(
	docRepo repository.DocumentRepository,
	chunkRepo repository.ChunkRepository,
	llmClient llm.LLMClient,
) DocumentService {
	return &documentService{
		docRepo:    docRepo,
		chunkRepo:  chunkRepo,
		llm:        llmClient,
		chunkerCfg: chunker.DefaultConfig(),
	}
}

func (s *documentService) Upload(ctx context.Context, userID uuid.UUID, file io.Reader, filename string, fileSize int64) (*model.Document, error) {
	if err := s.validateFile(filename, fileSize); err != nil {
		return nil, err
	}

	text, err := parser.Parse(file, filename)
	if err != nil {
		return nil, model.NewValidationError("Failed to parse file: " + err.Error())
	}

	if strings.TrimSpace(text) == "" {
		return nil, model.NewValidationError("File is empty or contains no extractable text")
	}

	chunks := chunker.Chunk(text, s.chunkerCfg)
	if len(chunks) == 0 {
		return nil, model.NewValidationError("File produced no chunks after processing")
	}

	chunkTexts := make([]string, len(chunks))
	copy(chunkTexts, chunks)

	embeddings, err := s.llm.Embed(ctx, chunkTexts)
	if err != nil {
		return nil, model.NewAppError(502, "embedding_error", "Failed to generate embeddings: "+err.Error())
	}

	if len(embeddings) != len(chunks) {
		return nil, model.NewAppError(502, "embedding_error", "Embedding count mismatch")
	}

	doc := &model.Document{
		ID:         uuid.New(),
		UserID:     userID,
		Filename:   filename,
		FileType:   strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), "."),
		FileSize:   fileSize,
		ChunkCount: len(chunks),
		CreatedAt:  time.Now(),
	}

	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, err
	}

	now := time.Now()
	modelChunks := make([]model.Chunk, len(chunks))
	for i, content := range chunks {
		modelChunks[i] = model.Chunk{
			ID:         uuid.New(),
			DocumentID: doc.ID,
			Content:    content,
			ChunkIndex: i,
			Embedding:  pgvector.NewVector(embeddings[i]),
			CreatedAt:  now,
		}
	}

	if err := s.chunkRepo.CreateBatch(ctx, modelChunks); err != nil {
		_ = s.docRepo.Delete(ctx, doc.ID)
		return nil, err
	}

	return doc, nil
}

func (s *documentService) List(ctx context.Context, userID uuid.UUID) ([]model.Document, error) {
	docs, err := s.docRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []model.Document{}
	}
	return docs, nil
}

func (s *documentService) Delete(ctx context.Context, userID uuid.UUID, docID uuid.UUID) error {
	doc, err := s.docRepo.GetByID(ctx, docID)
	if err != nil {
		return err
	}

	if doc.UserID != userID {
		return model.ErrForbidden
	}

	return s.docRepo.Delete(ctx, docID)
}

func (s *documentService) validateFile(filename string, fileSize int64) error {
	if filename == "" {
		return model.NewValidationError("Filename is required")
	}

	if fileSize <= 0 {
		return model.NewValidationError("File is empty")
	}

	if fileSize > maxFileSize {
		return model.NewValidationError("File size exceeds 10MB limit")
	}

	if !parser.IsSupportedExtension(filename) {
		return model.NewValidationError("Unsupported file type. Allowed: .txt, .md, .pdf")
	}

	return nil
}
