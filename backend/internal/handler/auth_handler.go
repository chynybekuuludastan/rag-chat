package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/dastanchynybek/rag-chat/backend/internal/service"
)

type AuthHandler struct {
	authService     service.AuthService
	refreshTokenTTL time.Duration
}

func NewAuthHandler(authService service.AuthService, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		refreshTokenTTL: refreshTTL,
	}
}

type registerRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepass123"`
}

type loginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepass123"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Registration data"
// @Success 201 {object} registerResponse
// @Failure 400 {object} errorResponse
// @Failure 409 {object} errorResponse "Email already exists"
// @Failure 422 {object} errorResponse "Validation error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := h.authService.Register(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate with email and password, returns access token and sets refresh token cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login credentials"
// @Success 200 {object} loginResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	tokens, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(h.refreshTokenTTL),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/api/auth",
	})

	return c.JSON(fiber.Map{
		"access_token": tokens.AccessToken,
		"expires_at":   tokens.ExpiresAt,
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Exchange refresh token (from cookie) for a new access token pair
// @Tags auth
// @Produce json
// @Success 200 {object} loginResponse
// @Failure 401 {object} errorResponse "Invalid or expired refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Missing refresh token")
	}

	tokens, err := h.authService.Refresh(c.Context(), refreshToken)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(h.refreshTokenTTL),
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/api/auth",
	})

	return c.JSON(fiber.Map{
		"access_token": tokens.AccessToken,
		"expires_at":   tokens.ExpiresAt,
	})
}
