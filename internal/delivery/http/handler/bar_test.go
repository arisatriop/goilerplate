package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/domain/bar"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubBarUsecase answers List without a database. The question these tests ask is what bytes
// the handler puts on the wire for a given result, which a real database only makes harder to
// arrange; the repository suite covers the query itself.
type stubBarUsecase struct {
	bar.Usecase

	bars   []*bar.Bar
	total  int64
	err    error
	filter *bar.Filter
}

func (s *stubBarUsecase) GetList(_ context.Context, filter *bar.Filter) ([]*bar.Bar, int64, error) {
	s.filter = filter
	if s.err != nil {
		return nil, 0, s.err
	}
	return s.bars, s.total, nil
}

func listBars(t *testing.T, usecase bar.Usecase, query string) (int, string) {
	t.Helper()

	app := fiber.New()
	app.Get("/bars", handler.NewBar(validator.New(), usecase).List)

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/bars?"+query, nil))
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, string(body)
}

// The exact bytes, because this is the endpoint the Swagger annotation describes and the one a
// new domain is copied from. It previously returned pagination.PaginatedResponse nested inside
// response.Success while documenting itself as response.PaginatedResponse — neither of which
// was the shape in .claude/rules/api-conventions.md.
func TestBarList_ReturnsTheDocumentedShape(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{
		bars:  []*bar.Bar{{ID: "1", Code: "EXP001", Bar: "first"}, {ID: "2", Code: "EXP002", Bar: "second"}},
		total: 42,
	}

	// Act
	status, body := listBars(t, usecase, "page=2&limit=10")

	// Assert
	require.Equal(t, fiber.StatusOK, status)
	assert.JSONEq(t, `{
		"success": true,
		"message": "Bars fetched successfully",
		"data": [
			{"id": "1", "code": "EXP001", "bar": "first"},
			{"id": "2", "code": "EXP002", "bar": "second"}
		],
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

func TestBarList_EmptyPageStillReturnsAnArray(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{bars: nil, total: 0}

	// Act
	status, body := listBars(t, usecase, "")

	// Assert
	require.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, `"data":[]`)
	assert.Contains(t, body, `"total":0`)
	assert.Contains(t, body, `"hasNext":false`)
}

func TestBarList_PassesThePagingThroughToTheUsecase(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{}

	// Act
	_, _ = listBars(t, usecase, "page=3&limit=25&keyword=beer")

	// Assert
	require.NotNil(t, usecase.filter)
	assert.Equal(t, "beer", usecase.filter.Keyword)
	assert.Equal(t, 3, usecase.filter.Pagination.Page)
	assert.Equal(t, 25, usecase.filter.Pagination.Limit)
	assert.Equal(t, 50, usecase.filter.Pagination.GetOffset())
}

// A client asking for a million rows gets the cap, not the million.
func TestBarList_ClampsAnOversizedLimit(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{}

	// Act
	_, body := listBars(t, usecase, "limit=1000000")

	// Assert
	require.NotNil(t, usecase.filter)
	assert.Equal(t, 100, usecase.filter.Pagination.Limit)
	assert.Contains(t, body, `"limit":100`)
}

func TestBarList_ReportsAUsecaseFailureAsAServerError(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{err: errors.New("database is on fire")}

	// Act
	status, body := listBars(t, usecase, "")

	// Assert
	assert.Equal(t, fiber.StatusInternalServerError, status)
	assert.Contains(t, body, `"success":false`)
	assert.NotContains(t, body, "database is on fire", "internal detail must not reach the client")
}
