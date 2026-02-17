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
	Question  string  `json:"question"`
	SessionID *string `json:"session_id,omitempty"`
}

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
