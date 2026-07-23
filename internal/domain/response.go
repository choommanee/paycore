package domain

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// APIResponse is the single envelope every endpoint returns.
type APIResponse struct {
	Success   bool        `json:"success"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
}

func requestID(c *fiber.Ctx) string {
	if id, ok := c.Locals("request_id").(string); ok && id != "" {
		return id
	}
	return uuid.NewString()
}

func Success(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true, Code: "OK", Message: "success",
		Data: data, RequestID: requestID(c), Timestamp: time.Now().UTC(),
	})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true, Code: "CREATED", Message: "created",
		Data: data, RequestID: requestID(c), Timestamp: time.Now().UTC(),
	})
}

func Error(c *fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(APIResponse{
		Success: false, Code: code, Message: msg,
		RequestID: requestID(c), Timestamp: time.Now().UTC(),
	})
}

// ErrorData is Error with a machine-parseable data payload (e.g. a
// field-oriented validation errors array). The envelope shape is unchanged.
func ErrorData(c *fiber.Ctx, status int, code, msg string, data interface{}) error {
	return c.Status(status).JSON(APIResponse{
		Success: false, Code: code, Message: msg, Data: data,
		RequestID: requestID(c), Timestamp: time.Now().UTC(),
	})
}
