package service

import (
	"context"
	"errors"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
	"github.com/dastanchynybek/rag-chat/backend/internal/repository"
)

type authService struct {
	userRepo        repository.UserRepository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (*model.User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, model.WrapInternal(err)
	}

	user := &model.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hash),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, model.ErrConflict) {
			return nil, model.NewAppError(409, "conflict", "User with this email already exists")
		}
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, model.NewAppError(401, "unauthorized", "Invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, model.NewAppError(401, "unauthorized", "Invalid email or password")
	}

	return s.generateTokenPair(user.ID, user.Email)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, model.ErrUnauthorized
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, model.ErrUnauthorized
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, model.ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, model.ErrUnauthorized
	}

	return s.generateTokenPair(user.ID, user.Email)
}

func (s *authService) generateTokenPair(userID uuid.UUID, email string) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTokenTTL)

	accessClaims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"exp":     accessExp.Unix(),
		"iat":     now.Unix(),
		"type":    "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, model.WrapInternal(err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     now.Add(s.refreshTokenTTL).Unix(),
		"iat":     now.Unix(),
		"type":    "refresh",
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshTokenObj.SignedString(s.jwtSecret)
	if err != nil {
		return nil, model.WrapInternal(err)
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresAt:    accessExp.Unix(),
	}, nil
}

func (s *authService) parseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func validateEmail(email string) error {
	if email == "" {
		return model.NewValidationError("Email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return model.NewValidationError("Invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return model.NewValidationError("Password must be at least 8 characters")
	}
	return nil
}
