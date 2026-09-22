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

// The list shape is asserted against the exact bytes. There used to be three descriptions of
// it and no two agreed — the conventions doc, response.Paginated, and pkg/pagination — and the
// reference handler's Swagger annotation described none of them. This is the contract now.
func TestPaginated_ExactShape(t *testing.T) {
	// Arrange
	items := []map[string]any{{"id": 1}, {"id": 2}}

	// Act
	status, body := send(t, func(c *fiber.Ctx) error {
		return response.Paginated(c, items, response.NewPagination(42, 2, 10),
			response.WithMessage("Bars fetched successfully"))
	})

	// Assert
	assert.Equal(t, fiber.StatusOK, status)
	assert.JSONEq(t, `{
		"success": true,
		"message": "Bars fetched successfully",
		"data": [{"id": 1}, {"id": 2}],
		"meta": {
			"page": 2,
			"limit": 10,
			"total": 42,
			"totalPages": 5,
			"hasNext": true,
			"hasPrev": true
		}
	}`, body)
}

// data stays a plain array. Nesting it under data.items — which is what the reference handler
// used to return — means every client unwraps one level before it can iterate.
func TestPaginated_DataIsAPlainArray(t *testing.T) {
	_, body := send(t, func(c *fiber.Ctx) error {
		return response.Paginated(c, []int{1, 2, 3}, response.NewPagination(3, 1, 10))
	})

	var payload struct {
		Data []int `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	assert.Equal(t, []int{1, 2, 3}, payload.Data)
}

func TestNewPagination_Counters(t *testing.T) {
	tests := []struct {
		name                     string
		total                    int64
		page, limit              int
		wantTotalPages           int
		wantHasNext, wantHasPrev bool
	}{
		{"first page of many", 42, 1, 10, 5, true, false},
		{"middle page", 42, 3, 10, 5, true, true},
		{"last page", 42, 5, 10, 5, false, true},
		{"exactly one full page", 10, 1, 10, 1, false, false},
		{"one row over a page boundary", 11, 1, 10, 2, true, false},
		{"empty result set", 0, 1, 10, 0, false, false},
		{"page past the end", 42, 9, 10, 5, false, true},
		// limit 0 would divide by zero; the parser never produces it, but a direct caller could.
		{"zero limit does not panic", 5, 1, 0, 5, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := response.NewPagination(tt.total, tt.page, tt.limit)

			assert.Equal(t, tt.wantTotalPages, page.TotalPages)
			assert.Equal(t, tt.wantHasNext, page.HasNext)
			assert.Equal(t, tt.wantHasPrev, page.HasPrev)
			assert.Equal(t, tt.total, page.Total)
		})
	}
}

// An empty page must marshal data as [], not null: a client iterating it should not need a nil
// check to tell "no results" from "no field".
func TestPaginated_EmptyPageIsAnEmptyArray(t *testing.T) {
	tests := []struct {
		name string
		data any
	}{
		{"nil slice", []int(nil)},
		{"empty slice", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, body := send(t, func(c *fiber.Ctx) error {
				return response.Paginated(c, tt.data, response.NewPagination(0, 1, 10))
			})

			assert.Contains(t, body, `"data":[]`)
			assert.NotContains(t, body, `"data":null`)
		})
	}
}

// Pagination is embedded by pointer, so a non-list response must not grow empty page counters.
func TestSuccess_CarriesNoPaginationKeys(t *testing.T) {
	_, body := send(t, func(c *fiber.Ctx) error {
		return response.Success(c, fiber.Map{"id": 1},
			response.WithMeta(&response.Meta{RequestID: "req-1"}))
	})

	for _, key := range []string{"page", "limit", "total", "totalPages", "hasNext", "hasPrev"} {
		assert.NotContains(t, body, `"`+key+`"`)
	}
}

// WithMeta on a list response sets the request ID without dropping the page counters.
func TestPaginated_MetaOptionKeepsThePageCounters(t *testing.T) {
	_, body := send(t, func(c *fiber.Ctx) error {
		return response.Paginated(c, []int{1}, response.NewPagination(1, 1, 10),
			response.WithMeta(&response.Meta{RequestID: "req-1"}))
	})

	var payload struct {
		Meta struct {
			RequestID string `json:"requestId"`
			Page      int    `json:"page"`
			Total     int64  `json:"total"`
		} `json:"meta"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	assert.Equal(t, "req-1", payload.Meta.RequestID)
	assert.Equal(t, 1, payload.Meta.Page)
	assert.Equal(t, int64(1), payload.Meta.Total)
}
