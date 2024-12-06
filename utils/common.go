package utils

import (
	"fmt"
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

func ErrorResponse(c *fiber.Ctx, statusCode int, message string, err error) error {
	_, file, line, _ := runtime.Caller(1)
	// Format the location info
	location := fmt.Sprintf("%s:%d", file, line)

	log.Printf("%s at %s: %v", message, location, err)

	return c.Status(statusCode).JSON(fiber.Map{
		"error":    message,
		"details":  err.Error(),
		"location": location,
	})
}
