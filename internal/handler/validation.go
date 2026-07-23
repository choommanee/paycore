package handler

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/yourco/payment-gateway/internal/domain"
)

// FieldError is a single machine-parseable validation failure. It uses the JSON
// field name (not the Go struct field) and the validator tag as a stable code so
// integrators can react programmatically instead of parsing English prose.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// validationDetails converts a go-playground/validator error into a stable,
// field-oriented slice. It emits the lower_snake_case wire field name and the
// failing tag as the code, never leaking internal Go struct/tag paths. A
// non-validation error returns a single generic entry so callers still get a
// well-formed envelope.
func validationDetails(err error) []FieldError {
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{{Field: "", Code: "invalid", Message: "request validation failed"}}
	}
	out := make([]FieldError, 0, len(verrs))
	for _, fe := range verrs {
		out = append(out, FieldError{
			Field:   toWireName(fe.Field()),
			Code:    fe.Tag(),
			Message: fieldMessage(fe),
		})
	}
	return out
}

// toWireName converts a Go exported field name to the lower_snake_case wire name
// used across the API's JSON tags (e.g. "SettlementCurrency" -> "settlement_currency").
func toWireName(goField string) string {
	var b strings.Builder
	for i, r := range goField {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// fieldMessage builds a short, human-readable message from the failing tag and
// its parameter without echoing Go type/namespace internals.
func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "len":
		return "must be exactly " + fe.Param() + " characters"
	case "min":
		return "must be at least " + fe.Param()
	case "max":
		return "must be at most " + fe.Param()
	case "gt":
		return "must be greater than " + fe.Param()
	case "gte":
		return "must be greater than or equal to " + fe.Param()
	case "oneof":
		return "must be one of: " + fe.Param()
	case "uuid", "uuid4":
		return "must be a valid UUID"
	default:
		if fe.Param() != "" {
			return "failed the '" + fe.Tag() + "=" + fe.Param() + "' constraint"
		}
		return "failed the '" + fe.Tag() + "' constraint"
	}
}

// validationErrorResponse writes the standard validation-failure envelope: a
// generic top-level message plus a machine-parseable data.errors array. Applied
// by every handler so integrators get one stable shape.
func validationErrorResponse(c *fiber.Ctx, err error) error {
	return domain.ErrorData(c, fiber.StatusBadRequest, "VALIDATION_ERROR",
		"request validation failed", map[string]interface{}{"errors": validationDetails(err)})
}
