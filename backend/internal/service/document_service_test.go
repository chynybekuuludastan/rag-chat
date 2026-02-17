package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
	"github.com/dastanchynybek/rag-chat/backend/internal/service/mocks"
)

func newTestDocumentService(t *testing.T) (*gomock.Controller, *mocks.MockDocumentRepository, *mocks.MockChunkRepository, *mocks.MockLLMClient, DocumentService) {
	ctrl := gomock.NewController(t)
	mockDocRepo := mocks.NewMockDocumentRepository(ctrl)
	mockChunkRepo := mocks.NewMockChunkRepository(ctrl)
	mockLLM := mocks.NewMockLLMClient(ctrl)
	svc := NewDocumentService(mockDocRepo, mockChunkRepo, mockLLM)
	return ctrl, mockDocRepo, mockChunkRepo, mockLLM, svc
}

func TestDocumentService_Upload_Success(t *testing.T) {
	ctrl, mockDocRepo, mockChunkRepo, mockLLM, svc := newTestDocumentService(t)
	defer ctrl.Finish()

	content := strings.Repeat("This is test content for chunking. ", 30)
	reader := strings.NewReader(content)
	userID := uuid.New()

	mockLLM.EXPECT().
		Embed(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, texts []string) ([][]float32, error) {
			result := make([][]float32, len(texts))
			for i := range texts {
				result[i] = make([]float32, 1536)
			}
			return result, nil
		})

	mockDocRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)

	mockChunkRepo.EXPECT().
		CreateBatch(gomock.Any(), gomock.Any()).
		Return(nil)

	doc, err := svc.Upload(context.Background(), userID, reader, "test.txt", int64(len(content)))

	require.NoError(t, err)
	assert.Equal(t, "test.txt", doc.Filename)
	assert.Equal(t, "txt", doc.FileType)
	assert.Greater(t, doc.ChunkCount, 0)
	assert.Equal(t, userID, doc.UserID)
}

func TestDocumentService_Upload_InvalidFileType(t *testing.T) {
	ctrl, _, _, _, svc := newTestDocumentService(t)
	defer ctrl.Finish()

	reader := strings.NewReader("content")

	doc, err := svc.Upload(context.Background(), uuid.New(), reader, "test.exe", 7)

	assert.Nil(t, doc)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.Code)
}

func TestDocumentService_Upload_EmptyFile(t *testing.T) {
	ctrl, _, _, _, svc := newTestDocumentService(t)
	defer ctrl.Finish()

	doc, err := svc.Upload(context.Background(), uuid.New(), strings.NewReader(""), "test.txt", 0)

	assert.Nil(t, doc)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.Code)
}

func TestDocumentService_Upload_FileTooLarge(t *testing.T) {
	ctrl, _, _, _, svc := newTestDocumentService(t)
	defer ctrl.Finish()

	doc, err := svc.Upload(context.Background(), uuid.New(), strings.NewReader("x"), "test.txt", 11*1024*1024)

	assert.Nil(t, doc)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.Code)
}

func TestDocumentService_Upload_EmbeddingAPIFailure(t *testing.T) {
	ctrl, _, _, mockLLM, svc := newTestDocumentService(t)
	defer ctrl.Finish()

	content := strings.Repeat("Test content for embedding failure. ", 30)

	mockLLM.EXPECT().
		Embed(gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	doc, err := svc.Upload(context.Background(), uuid.New(), strings.NewReader(content), "test.txt", int64(len(content)))

	assert.Nil(t, doc)
	require.Error(t, err)
}

func TestDocumentService_List_ReturnsOnlyUserDocs(t *testing.T) {
	ctrl, mockDocRepo, _, _, svc := newTestDocumentService(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedDocs := []model.Document{
		{ID: uuid.New(), UserID: userID, Filename: "doc1.txt", CreatedAt: time.Now()},
		{ID: uuid.New(), UserID: userID, Filename: "doc2.md", CreatedAt: time.Now()},
	}

	mockDocRepo.EXPECT().
		ListByUser(gomock.Any(), userID).
		Return(expectedDocs, nil)

	docs, err := svc.List(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.Equal(t, "doc1.txt", docs[0].Filename)
}

func TestDocumentService_Delete_NotOwner(t *testing.T) {
	ctrl, mockDocRepo, _, _, svc := newTestDocumentService(t)
	defer ctrl.Finish()

	ownerID := uuid.New()
	otherUserID := uuid.New()
	docID := uuid.New()

	mockDocRepo.EXPECT().
		GetByID(gomock.Any(), docID).
		Return(&model.Document{ID: docID, UserID: ownerID}, nil)

	err := svc.Delete(context.Background(), otherUserID, docID)

	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}
