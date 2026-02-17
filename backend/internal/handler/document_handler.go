package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/dastanchynybek/rag-chat/backend/internal/middleware"
	"github.com/dastanchynybek/rag-chat/backend/internal/service"
)

type DocumentHandler struct {
	docService service.DocumentService
}

func NewDocumentHandler(docService service.DocumentService) *DocumentHandler {
	return &DocumentHandler{docService: docService}
}

func (h *DocumentHandler) Upload(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "File is required")
	}

	f, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Failed to read file")
	}
	defer f.Close()

	doc, err := h.docService.Upload(c.Context(), userID, f, file.Filename, file.Size)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(doc)
}

func (h *DocumentHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	docs, err := h.docService.List(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"documents": docs,
	})
}

func (h *DocumentHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	docID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid document ID")
	}

	if err := h.docService.Delete(c.Context(), userID, docID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
