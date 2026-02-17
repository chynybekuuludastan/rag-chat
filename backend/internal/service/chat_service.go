package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
	"github.com/dastanchynybek/rag-chat/backend/internal/pkg/llm"
	"github.com/dastanchynybek/rag-chat/backend/internal/repository"
)

const (
	similarityThreshold = 0.3
	topKChunks          = 3
	maxQuestionLength   = 2000
)

const systemPrompt = `You are a helpful assistant that answers questions based on the provided documents.
Use ONLY the information from the context below. If the answer is not found
in the context, say so honestly. Always respond in the same language as the question.`

type chatService struct {
	chatRepo  repository.ChatRepository
	chunkRepo repository.ChunkRepository
	llm       llm.LLMClient
}

func NewChatService(
	chatRepo repository.ChatRepository,
	chunkRepo repository.ChunkRepository,
	llmClient llm.LLMClient,
) ChatService {
	return &chatService{
		chatRepo:  chatRepo,
		chunkRepo: chunkRepo,
		llm:       llmClient,
	}
}

func (s *chatService) Ask(ctx context.Context, userID uuid.UUID, sessionID *uuid.UUID, question string) (*uuid.UUID, <-chan ChatEvent, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, nil, model.NewValidationError("Question is required")
	}
	if len([]rune(question)) > maxQuestionLength {
		return nil, nil, model.NewValidationError(fmt.Sprintf("Question exceeds %d character limit", maxQuestionLength))
	}

	var sid uuid.UUID
	if sessionID != nil {
		session, err := s.chatRepo.GetSession(ctx, *sessionID)
		if err != nil {
			return nil, nil, err
		}
		if session.UserID != userID {
			return nil, nil, model.ErrForbidden
		}
		sid = *sessionID
	} else {
		sid = uuid.New()
		session := &model.ChatSession{
			ID:        sid,
			UserID:    userID,
			Title:     truncateTitle(question, 50),
			CreatedAt: time.Now(),
		}
		if err := s.chatRepo.CreateSession(ctx, session); err != nil {
			return nil, nil, err
		}
	}

	userMsg := &model.Message{
		ID:        uuid.New(),
		SessionID: sid,
		Role:      "user",
		Content:   question,
		CreatedAt: time.Now(),
	}
	if err := s.chatRepo.CreateMessage(ctx, userMsg); err != nil {
		return nil, nil, err
	}

	questionEmbedding, err := s.llm.Embed(ctx, []string{question})
	if err != nil {
		return nil, nil, model.NewAppError(502, "embedding_error", fmt.Sprintf("Failed to embed question: %v", err))
	}

	relevantChunks, err := s.chunkRepo.SearchSimilar(ctx, questionEmbedding[0], userID, topKChunks, similarityThreshold)
	if err != nil {
		return nil, nil, err
	}

	prompt := buildPrompt(relevantChunks, question)

	contentCh, errCh := s.llm.ChatStream(ctx, prompt)

	eventCh := make(chan ChatEvent, 64)

	go func() {
		defer close(eventCh)

		var fullResponse strings.Builder

		for content := range contentCh {
			fullResponse.WriteString(content)
			eventCh <- ChatEvent{
				Type:    "chunk",
				Content: content,
			}
		}

		if err := <-errCh; err != nil {
			eventCh <- ChatEvent{
				Type:    "error",
				Content: err.Error(),
			}
			return
		}

		sources := make([]SourceRef, len(relevantChunks))
		sourceChunkIDs := make([]uuid.UUID, len(relevantChunks))
		for i, c := range relevantChunks {
			sources[i] = SourceRef{
				ID:               c.ID,
				DocumentFilename: c.DocumentFilename,
				Content:          c.Content,
				Similarity:       c.Similarity,
			}
			sourceChunkIDs[i] = c.ID
		}

		if len(sources) > 0 {
			eventCh <- ChatEvent{
				Type:    "sources",
				Sources: sources,
			}
		}

		assistantMsg := &model.Message{
			ID:           uuid.New(),
			SessionID:    sid,
			Role:         "assistant",
			Content:      fullResponse.String(),
			SourceChunks: sourceChunkIDs,
			CreatedAt:    time.Now(),
		}
		_ = s.chatRepo.CreateMessage(context.Background(), assistantMsg)

		eventCh <- ChatEvent{Type: "done"}
	}()

	return &sid, eventCh, nil
}

func (s *chatService) GetHistory(ctx context.Context, userID uuid.UUID) ([]model.ChatSession, error) {
	sessions, err := s.chatRepo.ListSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if sessions == nil {
		sessions = []model.ChatSession{}
	}
	return sessions, nil
}

func (s *chatService) GetMessages(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) ([]MessageWithSources, error) {
	session, err := s.chatRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, model.ErrForbidden
	}

	messages, err := s.chatRepo.GetMessagesBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var allChunkIDs []uuid.UUID
	for _, msg := range messages {
		allChunkIDs = append(allChunkIDs, msg.SourceChunks...)
	}

	chunkMap := make(map[uuid.UUID]model.ChunkWithDocument)
	if len(allChunkIDs) > 0 {
		chunks, err := s.chunkRepo.GetByIDs(ctx, allChunkIDs)
		if err != nil {
			return nil, err
		}
		for _, c := range chunks {
			chunkMap[c.ID] = c
		}
	}

	result := make([]MessageWithSources, len(messages))
	for i, msg := range messages {
		result[i] = MessageWithSources{Message: msg}
		for _, chunkID := range msg.SourceChunks {
			if c, ok := chunkMap[chunkID]; ok {
				result[i].Sources = append(result[i].Sources, SourceRef{
					ID:               c.ID,
					DocumentFilename: c.DocumentFilename,
					Content:          c.Content,
				})
			}
		}
	}

	if result == nil {
		result = []MessageWithSources{}
	}
	return result, nil
}

func buildPrompt(chunks []model.ChunkWithDocument, question string) []llm.ChatMessage {
	var contextBuilder strings.Builder
	if len(chunks) > 0 {
		contextBuilder.WriteString("\n\nContext:\n---\n")
		for _, c := range chunks {
			contextBuilder.WriteString(fmt.Sprintf("[Source: %s, chunk %d]\n%s\n\n",
				c.DocumentFilename, c.ChunkIndex, c.Content))
		}
		contextBuilder.WriteString("---")
	} else {
		contextBuilder.WriteString("\n\nNo relevant documents were found for this question.")
	}

	return []llm.ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt + contextBuilder.String(),
		},
		{
			Role:    "user",
			Content: question,
		},
	}
}

func truncateTitle(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// Ensure chatService satisfies the interface at compile time.
var _ ChatService = (*chatService)(nil)

// Ensure errors are properly handled - compile-time check.
var _ error = (*model.AppError)(nil)
