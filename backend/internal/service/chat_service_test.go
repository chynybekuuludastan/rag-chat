package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
	"github.com/dastanchynybek/rag-chat/backend/internal/service/mocks"
)

func newTestChatService(t *testing.T) (*gomock.Controller, *mocks.MockChatRepository, *mocks.MockChunkRepository, *mocks.MockLLMClient, ChatService) {
	ctrl := gomock.NewController(t)
	mockChatRepo := mocks.NewMockChatRepository(ctrl)
	mockChunkRepo := mocks.NewMockChunkRepository(ctrl)
	mockLLM := mocks.NewMockLLMClient(ctrl)
	svc := NewChatService(mockChatRepo, mockChunkRepo, mockLLM)
	return ctrl, mockChatRepo, mockChunkRepo, mockLLM, svc
}

func TestChatService_Ask_Success(t *testing.T) {
	ctrl, mockChatRepo, mockChunkRepo, mockLLM, svc := newTestChatService(t)
	defer ctrl.Finish()

	userID := uuid.New()
	question := "What is the revenue?"

	// Create session
	mockChatRepo.EXPECT().
		CreateSession(gomock.Any(), gomock.Any()).
		Return(nil)

	// Save user message
	mockChatRepo.EXPECT().
		CreateMessage(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1) // user message

	// Embed question
	mockLLM.EXPECT().
		Embed(gomock.Any(), []string{question}).
		Return([][]float32{make([]float32, 1536)}, nil)

	// Search similar chunks
	mockChunkRepo.EXPECT().
		SearchSimilar(gomock.Any(), gomock.Any(), userID, topKChunks, similarityThreshold).
		Return([]model.ChunkWithDocument{
			{
				Chunk:            model.Chunk{ID: uuid.New(), Content: "Revenue was $1M", ChunkIndex: 0},
				DocumentFilename: "report.pdf",
				Similarity:       0.85,
			},
		}, nil)

	// Stream response
	contentCh := make(chan string, 3)
	errCh := make(chan error, 1)
	contentCh <- "According to"
	contentCh <- " the report,"
	contentCh <- " revenue was $1M."
	close(contentCh)
	close(errCh)

	mockLLM.EXPECT().
		ChatStream(gomock.Any(), gomock.Any()).
		Return(contentCh, errCh)

	// Save assistant message (called from goroutine)
	mockChatRepo.EXPECT().
		CreateMessage(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1) // assistant message

	sid, eventCh, err := svc.Ask(context.Background(), userID, nil, question)

	require.NoError(t, err)
	require.NotNil(t, sid)

	var events []ChatEvent
	for event := range eventCh {
		events = append(events, event)
	}

	// Should have: 3 chunks + 1 sources + 1 done = 5 events
	require.GreaterOrEqual(t, len(events), 3)

	// Verify we got content chunks
	contentEvents := 0
	for _, e := range events {
		if e.Type == "chunk" {
			contentEvents++
		}
	}
	assert.Equal(t, 3, contentEvents)

	// Verify last event is done
	assert.Equal(t, "done", events[len(events)-1].Type)
}

func TestChatService_Ask_NoRelevantChunks(t *testing.T) {
	ctrl, mockChatRepo, mockChunkRepo, mockLLM, svc := newTestChatService(t)
	defer ctrl.Finish()

	userID := uuid.New()

	mockChatRepo.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil)
	mockChatRepo.EXPECT().CreateMessage(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	mockLLM.EXPECT().
		Embed(gomock.Any(), gomock.Any()).
		Return([][]float32{make([]float32, 1536)}, nil)

	// No relevant chunks found
	mockChunkRepo.EXPECT().
		SearchSimilar(gomock.Any(), gomock.Any(), userID, topKChunks, similarityThreshold).
		Return([]model.ChunkWithDocument{}, nil)

	contentCh := make(chan string, 1)
	errCh := make(chan error, 1)
	contentCh <- "I could not find relevant information in the documents."
	close(contentCh)
	close(errCh)

	mockLLM.EXPECT().ChatStream(gomock.Any(), gomock.Any()).Return(contentCh, errCh)
	mockChatRepo.EXPECT().CreateMessage(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	sid, eventCh, err := svc.Ask(context.Background(), userID, nil, "Unknown question?")

	require.NoError(t, err)
	require.NotNil(t, sid)

	var hasSourcesEvent bool
	for event := range eventCh {
		if event.Type == "sources" {
			hasSourcesEvent = true
		}
	}
	// No sources should be emitted when no relevant chunks
	assert.False(t, hasSourcesEvent)
}

func TestChatService_Ask_EmptyQuestion(t *testing.T) {
	ctrl, _, _, _, svc := newTestChatService(t)
	defer ctrl.Finish()

	sid, eventCh, err := svc.Ask(context.Background(), uuid.New(), nil, "   ")

	assert.Nil(t, sid)
	assert.Nil(t, eventCh)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.Code)
}

func TestChatService_Ask_SessionNotOwned(t *testing.T) {
	ctrl, mockChatRepo, _, _, svc := newTestChatService(t)
	defer ctrl.Finish()

	userID := uuid.New()
	otherUserID := uuid.New()
	sessionID := uuid.New()

	mockChatRepo.EXPECT().
		GetSession(gomock.Any(), sessionID).
		Return(&model.ChatSession{
			ID:     sessionID,
			UserID: otherUserID,
		}, nil)

	sid, eventCh, err := svc.Ask(context.Background(), userID, &sessionID, "Question?")

	assert.Nil(t, sid)
	assert.Nil(t, eventCh)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}

func TestChatService_GetHistory_Empty(t *testing.T) {
	ctrl, mockChatRepo, _, _, svc := newTestChatService(t)
	defer ctrl.Finish()

	userID := uuid.New()

	mockChatRepo.EXPECT().
		ListSessionsByUser(gomock.Any(), userID).
		Return(nil, nil)

	sessions, err := svc.GetHistory(context.Background(), userID)

	require.NoError(t, err)
	assert.NotNil(t, sessions) // should return empty slice, not nil
	assert.Len(t, sessions, 0)
}

func TestChatService_GetMessages_NotOwned(t *testing.T) {
	ctrl, mockChatRepo, _, _, svc := newTestChatService(t)
	defer ctrl.Finish()

	userID := uuid.New()
	otherUserID := uuid.New()
	sessionID := uuid.New()

	mockChatRepo.EXPECT().
		GetSession(gomock.Any(), sessionID).
		Return(&model.ChatSession{
			ID:        sessionID,
			UserID:    otherUserID,
			Title:     "Other user's chat",
			CreatedAt: time.Now(),
		}, nil)

	messages, err := svc.GetMessages(context.Background(), userID, sessionID)

	assert.Nil(t, messages)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}
