package handler

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/dastanchynybek/rag-chat/backend/internal/middleware"
	"github.com/dastanchynybek/rag-chat/backend/internal/service"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

type askRequest struct {
	Question  string  `json:"question" example:"What does the report say about revenue?"`
	SessionID *string `json:"session_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Ask godoc
// @Summary Ask a question (RAG)
// @Description Send a question to the RAG pipeline. Embeds the question, searches relevant chunks, and streams LLM response via SSE.
// @Tags chat
// @Accept json
// @Produce text/event-stream
// @Param request body askRequest true "Question and optional session ID"
// @Success 200 {string} string "SSE stream with ChatEvent objects"
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 422 {object} errorResponse "Empty question or exceeds 2000 chars"
// @Failure 502 {object} errorResponse "Embedding API failure"
// @Security BearerAuth
// @Router /chat [post]
func (h *ChatHandler) Ask(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req askRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	var sessionID *uuid.UUID
	if req.SessionID != nil && *req.SessionID != "" {
		id, err := uuid.Parse(*req.SessionID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid session ID")
		}
		sessionID = &id
	}

	sid, eventCh, err := h.chatService.Ask(c.Context(), userID, sessionID, req.Question)
	if err != nil {
		return err
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Session-ID", sid.String())

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for event := range eventCh {
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()
		}
	})

	return nil
}

// GetHistory godoc
// @Summary Get chat history
// @Description List all chat sessions for the current user
// @Tags chat
// @Produce json
// @Success 200 {object} sessionListResponse
// @Failure 401 {object} errorResponse
// @Security BearerAuth
// @Router /chat/history [get]
func (h *ChatHandler) GetHistory(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	sessions, err := h.chatService.GetHistory(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
	})
}

// GetMessages godoc
// @Summary Get messages for a chat session
// @Description Get all messages in a specific chat session. Only the session owner can access.
// @Tags chat
// @Produce json
// @Param sessionId path string true "Chat session UUID"
// @Success 200 {object} messageListResponse
// @Failure 400 {object} errorResponse "Invalid session ID"
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse "Not the session owner"
// @Failure 404 {object} errorResponse
// @Security BearerAuth
// @Router /chat/history/{sessionId} [get]
func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	sessionID, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid session ID")
	}

	messages, err := h.chatService.GetMessages(c.Context(), userID, sessionID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"messages": messages,
	})
}
