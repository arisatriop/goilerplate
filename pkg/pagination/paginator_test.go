package pagination_test

import (
	"net/http/httptest"
	"testing"

	"goilerplate/pkg/pagination"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseQuery runs ParsePagination against a real request, because it reads the query string,
// the form values and the body — behaviour a struct literal would not exercise.
func parseQuery(t *testing.T, query string) *pagination.PaginationRequest {
	t.Helper()

	app := fiber.New()
	var parsed *pagination.PaginationRequest
	app.Get("/", func(c *fiber.Ctx) error {
		parsed = pagination.ParsePagination(c)
		return c.SendStatus(fiber.StatusOK)
	})

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/?"+query, nil))
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.NotNil(t, parsed)
	return parsed
}

func TestParsePagination_Boundaries(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{"defaults when absent", "", 1, 10},
		{"explicit values", "page=3&limit=25", 3, 25},
		{"page zero falls back to the default", "page=0", 1, 10},
		{"negative page falls back to the default", "page=-4", 1, 10},
		{"limit zero falls back to the default", "limit=0", 1, 10},
		{"negative limit falls back to the default", "limit=-1", 1, 10},
		{"non-numeric values are ignored", "page=abc&limit=xyz", 1, 10},
		// Without a cap, ?limit=1000000 is a request for a million rows and the database pays
		// for refusing it.
		{"limit above the maximum is clamped", "limit=1000000", 1, 100},
		{"limit exactly at the maximum is kept", "limit=100", 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			parsed := parseQuery(t, tt.query)

			// Assert
			assert.Equal(t, tt.wantPage, parsed.Page)
			assert.Equal(t, tt.wantLimit, parsed.Limit)
		})
	}
}

func TestPaginationRequest_Offset(t *testing.T) {
	tests := []struct {
		page, limit, want int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 25, 50},
	}

	for _, tt := range tests {
		req := &pagination.PaginationRequest{Page: tt.page, Limit: tt.limit}
		assert.Equal(t, tt.want, req.GetOffset())
		assert.Equal(t, tt.limit, req.GetLimit())
	}
}

// A zero MaxLimit means "no cap", so a caller with its own config is not forced into one.
func TestPaginationRequest_ValidateWithoutAMaximum(t *testing.T) {
	req := &pagination.PaginationRequest{Page: 1, Limit: 5000}
	req.Validate(pagination.PaginationConfig{DefaultPage: 1, DefaultLimit: 10})

	assert.Equal(t, 5000, req.Limit)
}
