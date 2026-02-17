package middleware

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/dastanchynybek/rag-chat/backend/internal/model"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var appErr *model.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Code).JSON(fiber.Map{
			"error":   appErr.Err,
			"message": appErr.Message,
		})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"error":   "request_error",
			"message": fiberErr.Message,
		})
	}

	log.Printf("Unhandled error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":   "internal_error",
		"message": "Something went wrong",
	})
}
