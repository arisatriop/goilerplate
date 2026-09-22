package response_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"regexp"
	"testing"

	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// send runs one handler through a real Fiber app and returns the status and the raw body, so
// the assertions are made against the bytes a client would receive rather than a struct.
func send(t *testing.T, handler fiber.Handler) (int, string) {
	t.Helper()

	app := fiber.New()
	app.Get("/", handler)

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, string(body)
}

func TestEnvelope_SuccessHelpers(t *testing.T) {
	tests := []struct {
		name    string
		handler fiber.Handler
		status  int
		message string
	}{
		{
			name:    "Success",
			handler: func(c *fiber.Ctx) error { return response.Success(c, fiber.Map{"id": 1}) },
			status:  fiber.StatusOK,
			message: "Success",
		},
		{
			name:    "Created",
			handler: func(c *fiber.Ctx) error { return response.Created(c, fiber.Map{"id": 1}) },
			status:  fiber.StatusCreated,
			message: "Resource created successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			status, body := send(t, tt.handler)

			// Assert
			assert.Equal(t, tt.status, status)

			var payload struct {
				Success bool           `json:"success"`
				Message string         `json:"message"`
				Data    map[string]any `json:"data"`
			}
			require.NoError(t, json.Unmarshal([]byte(body), &payload))
			assert.True(t, payload.Success)
			assert.Equal(t, tt.message, payload.Message)
			assert.Equal(t, float64(1), payload.Data["id"])
		})
	}
}

func TestEnvelope_ErrorHelpers(t *testing.T) {
	tests := []struct {
		name    string
		handler fiber.Handler
		status  int
	}{
		{"BadRequest", func(c *fiber.Ctx) error { return response.BadRequest(c, "bad", nil) }, fiber.StatusBadRequest},
		{"Unauthorized", func(c *fiber.Ctx) error { return response.Unauthorized(c, "") }, fiber.StatusUnauthorized},
		{"Forbidden", func(c *fiber.Ctx) error { return response.Forbidden(c, "") }, fiber.StatusForbidden},
		{"NotFound", func(c *fiber.Ctx) error { return response.NotFound(c, "") }, fiber.StatusNotFound},
		{"Conflict", func(c *fiber.Ctx) error { return response.Conflict(c, "taken", nil) }, fiber.StatusConflict},
		{"UnprocessableEntity", func(c *fiber.Ctx) error { return response.UnprocessableEntity(c, "no", nil) }, fiber.StatusUnprocessableEntity},
		{"TooManyRequests", func(c *fiber.Ctx) error { return response.TooManyRequests(c, "") }, fiber.StatusTooManyRequests},
		{"InternalServerError", func(c *fiber.Ctx) error { return response.InternalServerError(c, "") }, fiber.StatusInternalServerError},
		{"ValidationError", func(c *fiber.Ctx) error { return response.ValidationError(c, nil) }, fiber.StatusBadRequest},
		{"CustomError", func(c *fiber.Ctx) error { return response.CustomError(c, fiber.StatusTeapot, "nope", nil) }, fiber.StatusTeapot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			status, body := send(t, tt.handler)

			// Assert
			assert.Equal(t, tt.status, status)

			var payload struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}
			require.NoError(t, json.Unmarshal([]byte(body), &payload))
			assert.False(t, payload.Success, "an error envelope must never report success")
			assert.NotEmpty(t, payload.Message, "every error carries a message, default or given")
		})
	}
}

// The default messages exist so a handler can pass "" and still produce something a client can
// show. A regression here turns an empty string into the user-visible text.
func TestEnvelope_DefaultMessages(t *testing.T) {
	tests := []struct {
		name    string
		handler fiber.Handler
		message string
	}{
		{"Unauthorized", func(c *fiber.Ctx) error { return response.Unauthorized(c, "") }, "Unauthorized"},
		{"Forbidden", func(c *fiber.Ctx) error { return response.Forbidden(c, "") }, "Forbidden"},
		{"NotFound", func(c *fiber.Ctx) error { return response.NotFound(c, "") }, "Not found"},
		{"ValidationError", func(c *fiber.Ctx) error { return response.ValidationError(c, nil) }, "Validation failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, body := send(t, tt.handler)
			assert.Contains(t, body, tt.message)
		})
	}
}

// An explicit message must win over the default.
func TestEnvelope_GivenMessageOverridesTheDefault(t *testing.T) {
	_, body := send(t, func(c *fiber.Ctx) error {
		return response.Unauthorized(c, "Session expired")
	})
	assert.Contains(t, body, "Session expired")
	assert.NotContains(t, body, `"message":"Unauthorized"`)
}

func TestEnvelope_WithMessageAndWithMeta(t *testing.T) {
	// Act
	_, body := send(t, func(c *fiber.Ctx) error {
		return response.Success(c, nil,
			response.WithMessage("Profile loaded"),
			response.WithMeta(&response.Meta{RequestID: "req-1", Timestamp: "2026-01-01T00:00:00Z"}))
	})

	// Assert
	var payload struct {
		Message string `json:"message"`
		Meta    struct {
			RequestID string `json:"requestId"`
			Timestamp string `json:"timestamp"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	assert.Equal(t, "Profile loaded", payload.Message)
	assert.Equal(t, "req-1", payload.Meta.RequestID)
	assert.Equal(t, "2026-01-01T00:00:00Z", payload.Meta.Timestamp)
}

// Meta used to emit request_id while every DTO around it emitted camelCase, so one response
// body carried two conventions. This walks the whole marshalled envelope so the next field
// added anywhere in it cannot reintroduce the split.
func TestEnvelope_EveryKeyIsCamelCase(t *testing.T) {
	// Arrange
	_, body := send(t, func(c *fiber.Ctx) error {
		return response.Success(c,
			fiber.Map{"userId": 1, "lastLoginAt": "2026-01-01T00:00:00Z"},
			response.WithMeta(&response.Meta{
				Message:   "ok",
				RequestID: "req-1",
				Timestamp: "2026-01-01T00:00:00Z",
			}))
	})

	// Act
	var payload any
	require.NoError(t, json.Unmarshal([]byte(body), &payload))

	// Assert
	for _, key := range jsonKeys(payload) {
		assert.Regexp(t, regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`), key,
			"response keys are camelCase; %q is not", key)
	}
}

// jsonKeys collects every object key in a decoded JSON document, at any depth.
func jsonKeys(value any) []string {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key, nested := range typed {
			keys = append(keys, key)
			keys = append(keys, jsonKeys(nested)...)
		}
		return keys
	case []any:
		var keys []string
		for _, nested := range typed {
			keys = append(keys, jsonKeys(nested)...)
		}
		return keys
	default:
		return nil
	}
}
