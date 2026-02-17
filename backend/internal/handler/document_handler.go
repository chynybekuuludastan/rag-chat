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

// Upload godoc
// @Summary Upload a document
// @Description Upload a text document (.txt, .md, .pdf) for RAG processing. The file will be chunked and embedded.
// @Tags documents
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Document file (.txt, .md, .pdf, max 10MB)"
// @Success 201 {object} documentResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 422 {object} errorResponse "Invalid file type or empty file"
// @Failure 502 {object} errorResponse "Embedding API failure"
// @Security BearerAuth
// @Router /documents [post]
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

// List godoc
// @Summary List uploaded documents
// @Description Get all documents uploaded by the current user
// @Tags documents
// @Produce json
// @Success 200 {object} documentListResponse
// @Failure 401 {object} errorResponse
// @Security BearerAuth
// @Router /documents [get]
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

// Delete godoc
// @Summary Delete a document
// @Description Delete a document and all associated chunks. Only the owner can delete.
// @Tags documents
// @Produce json
// @Param id path string true "Document UUID"
// @Success 204 "No Content"
// @Failure 400 {object} errorResponse "Invalid document ID"
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse "Not the document owner"
// @Failure 404 {object} errorResponse
// @Security BearerAuth
// @Router /documents/{id} [delete]
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
