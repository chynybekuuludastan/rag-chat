package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
	"github.com/dastanchynybek/rag-chat/backend/internal/service/mocks"
)

const testJWTSecret = "test-secret-key-must-be-at-least-32-characters-long"

func newTestAuthService(t *testing.T) (*gomock.Controller, *mocks.MockUserRepository, AuthService) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewAuthService(mockUserRepo, testJWTSecret, 15*time.Minute, 7*24*time.Hour)
	return ctrl, mockUserRepo, svc
}

func TestAuthService_Register_Success(t *testing.T) {
	ctrl, mockUserRepo, svc := newTestAuthService(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, user *model.User) error {
			assert.NotEqual(t, uuid.Nil, user.ID)
			assert.Equal(t, "test@example.com", user.Email)
			assert.NotEmpty(t, user.Password)
			assert.NotEqual(t, "password123", user.Password) // should be hashed
			return nil
		})

	user, err := svc.Register(context.Background(), "test@example.com", "password123")

	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	ctrl, mockUserRepo, svc := newTestAuthService(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(model.ErrConflict)

	user, err := svc.Register(context.Background(), "existing@example.com", "password123")

	assert.Nil(t, user)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	ctrl, _, svc := newTestAuthService(t)
	defer ctrl.Finish()

	user, err := svc.Register(context.Background(), "test@example.com", "short")

	assert.Nil(t, user)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.Code)
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	ctrl, _, svc := newTestAuthService(t)
	defer ctrl.Finish()

	user, err := svc.Register(context.Background(), "not-an-email", "password123")

	assert.Nil(t, user)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 422, appErr.Code)
}

func TestAuthService_Login_Success(t *testing.T) {
	ctrl, mockUserRepo, svc := newTestAuthService(t)
	defer ctrl.Finish()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	mockUserRepo.EXPECT().
		GetByEmail(gomock.Any(), "test@example.com").
		Return(&model.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Password:  string(hashed),
			CreatedAt: time.Now(),
		}, nil)

	tokens, err := svc.Login(context.Background(), "test@example.com", "password123")

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Greater(t, tokens.ExpiresAt, time.Now().Unix())
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	ctrl, mockUserRepo, svc := newTestAuthService(t)
	defer ctrl.Finish()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), 12)
	mockUserRepo.EXPECT().
		GetByEmail(gomock.Any(), "test@example.com").
		Return(&model.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			Password: string(hashed),
		}, nil)

	tokens, err := svc.Login(context.Background(), "test@example.com", "wrong-password")

	assert.Nil(t, tokens)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Code)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	ctrl, mockUserRepo, svc := newTestAuthService(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().
		GetByEmail(gomock.Any(), "nonexistent@example.com").
		Return(nil, model.ErrNotFound)

	tokens, err := svc.Login(context.Background(), "nonexistent@example.com", "password123")

	assert.Nil(t, tokens)
	require.Error(t, err)
	var appErr *model.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Code)
}

func TestAuthService_Refresh_ExpiredToken(t *testing.T) {
	ctrl, _, svc := newTestAuthService(t)
	defer ctrl.Finish()

	tokens, err := svc.Refresh(context.Background(), "invalid-token-string")

	assert.Nil(t, tokens)
	require.Error(t, err)
}
