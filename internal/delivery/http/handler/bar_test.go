package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
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
	single *bar.Bar
	total  int64
	err    error

	filter  *bar.Filter
	created *bar.Bar
	updated *bar.Bar
	deleted *bar.Bar
	gotID   string
}

func (s *stubBarUsecase) GetList(_ context.Context, filter *bar.Filter) ([]*bar.Bar, int64, error) {
	s.filter = filter
	if s.err != nil {
		return nil, 0, s.err
	}
	return s.bars, s.total, nil
}

func (s *stubBarUsecase) Create(_ context.Context, entity *bar.Bar) (*bar.Bar, error) {
	s.created = entity
	if s.err != nil {
		return nil, s.err
	}
	return entity, nil
}

func (s *stubBarUsecase) Update(_ context.Context, entity *bar.Bar) (*bar.Bar, error) {
	s.updated = entity
	if s.err != nil {
		return nil, s.err
	}
	return entity, nil
}

func (s *stubBarUsecase) Delete(_ context.Context, entity *bar.Bar) error {
	s.deleted = entity
	return s.err
}

func (s *stubBarUsecase) GetByID(_ context.Context, id string) (*bar.Bar, error) {
	s.gotID = id
	if s.err != nil {
		return nil, s.err
	}
	return s.single, nil
}

// call routes one request through a real Fiber app against the whole CRUD surface, so the
// route parameters and the body parser are exercised rather than stubbed.
func call(t *testing.T, usecase bar.Usecase, method, target, body string) (int, string) {
	t.Helper()

	h := handler.NewBar(validator.New(), usecase)
	app := fiber.New()
	app.Post("/bars", h.Create)
	app.Put("/bars/:id", h.Update)
	app.Delete("/bars/:id", h.Delete)
	app.Get("/bars/:id", h.Get)

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}

	res, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, string(raw)
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

// ── Create ───────────────────────────────────────────────────────────────────

func TestBarCreate_Returns201AndPassesTheEntityThrough(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{}

	// Act
	status, body := call(t, usecase, fiber.MethodPost, "/bars", `{"code":"EXP001","bar":"first"}`)

	// Assert
	assert.Equal(t, fiber.StatusCreated, status)
	assert.JSONEq(t, `{"success":true,"message":"Bar created successfully"}`, body)
	require.NotNil(t, usecase.created)
	assert.Equal(t, "EXP001", usecase.created.Code)
	assert.Equal(t, "first", usecase.created.Bar)
}

func TestBarCreate_RejectsBadInput(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		reason string
	}{
		{"malformed json", `{"code":`, "the body parser fails before validation"},
		{"missing code", `{"bar":"first"}`, "code is required"},
		{"code shorter than three", `{"code":"EX","bar":"first"}`, "gte=3"},
		{"missing bar", `{"code":"EXP001"}`, "bar is required"},
		{"empty object", `{}`, "both fields are required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			usecase := &stubBarUsecase{}

			// Act
			status, body := call(t, usecase, fiber.MethodPost, "/bars", tt.body)

			// Assert
			assert.Equal(t, fiber.StatusBadRequest, status, tt.reason)
			assert.Contains(t, body, `"success":false`)
			assert.Nil(t, usecase.created, "invalid input must never reach the use case")
		})
	}
}

// A domain error carries its own status code. A duplicate code is a 409, not a 500, and the
// handler must not flatten the difference.
func TestBarCreate_PropagatesTheDomainStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"duplicate code", bar.ErrCodeAlreadyExists, fiber.StatusConflict},
		{"not found", bar.ErrNotFound, fiber.StatusNotFound},
		{"cannot be deleted", bar.ErrCannotBeDeleted, fiber.StatusForbidden},
		{"already deleted", bar.ErrAlreadyDeleted, fiber.StatusGone},
		{"anything else", errors.New("connection reset"), fiber.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			usecase := &stubBarUsecase{err: tt.err}

			// Act
			status, body := call(t, usecase, fiber.MethodPost, "/bars", `{"code":"EXP001","bar":"first"}`)

			// Assert
			assert.Equal(t, tt.want, status)
			assert.Contains(t, body, `"success":false`)
		})
	}
}

// An unexpected error must not describe itself to the caller.
func TestBarCreate_ServerErrorLeaksNothing(t *testing.T) {
	usecase := &stubBarUsecase{err: errors.New("pq: relation \"bars\" does not exist")}

	_, body := call(t, usecase, fiber.MethodPost, "/bars", `{"code":"EXP001","bar":"first"}`)

	assert.NotContains(t, body, "relation")
	assert.NotContains(t, body, "pq:")
}

// ── Update ───────────────────────────────────────────────────────────────────

func TestBarUpdate_Returns200AndCarriesTheRouteID(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{}

	// Act
	status, body := call(t, usecase, fiber.MethodPut, "/bars/abc-123", `{"code":"EXP002","bar":"second"}`)

	// Assert
	assert.Equal(t, fiber.StatusOK, status)
	assert.JSONEq(t, `{"success":true,"message":"Bar updated successfully"}`, body)
	require.NotNil(t, usecase.updated)
	assert.Equal(t, "abc-123", usecase.updated.ID, "the id comes from the path, not the body")
	assert.Equal(t, "EXP002", usecase.updated.Code)
}

func TestBarUpdate_RejectsBadInput(t *testing.T) {
	usecase := &stubBarUsecase{}

	status, _ := call(t, usecase, fiber.MethodPut, "/bars/abc-123", `{"bar":"second"}`)

	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Nil(t, usecase.updated)
}

func TestBarUpdate_MissingRowIsA404(t *testing.T) {
	usecase := &stubBarUsecase{err: bar.ErrNotFound}

	status, _ := call(t, usecase, fiber.MethodPut, "/bars/missing", `{"code":"EXP002","bar":"second"}`)

	assert.Equal(t, fiber.StatusNotFound, status)
}

// ── Delete ───────────────────────────────────────────────────────────────────

// 204 carries no body by definition, so the envelope helpers must not write one.
func TestBarDelete_Returns204WithNoBody(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{}

	// Act
	status, body := call(t, usecase, fiber.MethodDelete, "/bars/abc-123", "")

	// Assert
	assert.Equal(t, fiber.StatusNoContent, status)
	assert.Empty(t, body)
	require.NotNil(t, usecase.deleted)
	assert.Equal(t, "abc-123", usecase.deleted.ID)
}

func TestBarDelete_MissingRowIsA404(t *testing.T) {
	usecase := &stubBarUsecase{err: bar.ErrNotFound}

	status, body := call(t, usecase, fiber.MethodDelete, "/bars/missing", "")

	assert.Equal(t, fiber.StatusNotFound, status)
	assert.Contains(t, body, `"success":false`)
}

// ── Get ──────────────────────────────────────────────────────────────────────

func TestBarGet_ReturnsTheDTOShape(t *testing.T) {
	// Arrange
	usecase := &stubBarUsecase{single: &bar.Bar{ID: "abc-123", Code: "EXP001", Bar: "first"}}

	// Act
	status, body := call(t, usecase, fiber.MethodGet, "/bars/abc-123", "")

	// Assert
	assert.Equal(t, fiber.StatusOK, status)
	assert.JSONEq(t, `{
		"success": true,
		"message": "Bar fetched successfully",
		"data": {"id": "abc-123", "code": "EXP001", "bar": "first"}
	}`, body)
	assert.Equal(t, "abc-123", usecase.gotID)
}

func TestBarGet_MissingRowIsA404(t *testing.T) {
	usecase := &stubBarUsecase{err: bar.ErrNotFound}

	status, _ := call(t, usecase, fiber.MethodGet, "/bars/missing", "")

	assert.Equal(t, fiber.StatusNotFound, status)
}
