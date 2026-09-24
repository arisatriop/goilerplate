package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goilerplate/internal/delivery/http/handler"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// passwordUsecase records whether ChangePassword was reached; any other method panics.
type passwordUsecase struct {
	auth.Usecase
	called bool
}

func (u *passwordUsecase) ChangePassword(context.Context, string, string, string, string) error {
	u.called = true
	return nil
}

// A rejected password used to come straight back in the 400 — and, because this route's response
// bodies were logged, into the logs as well. The response must name the field and say nothing of
// what was in it.
func TestAuthChangePassword_RejectionDoesNotEchoThePassword(t *testing.T) {
	// Arrange
	const current, rejected = "the-current-password", "hunter2"
	usecase := &passwordUsecase{}
	h := handler.NewAuth(nil, response.NewValidator(), nil, usecase, nil)

	app := fiber.New()
	app.Put("/users/me/password", func(ctx *fiber.Ctx) error {
		ctx.Locals(string(constants.ContextKeyUserID), "u1")
		ctx.Locals(string(constants.ContextKeySessionID), "s1")
		return ctx.Next()
	}, h.ChangePassword)

	req := httptest.NewRequest(fiber.MethodPut, "/users/me/password",
		strings.NewReader(`{"currentPassword":"`+current+`","newPassword":"`+rejected+`"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// Act
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.False(t, usecase.called, "a rejected request must not reach the use case")
	assert.NotContains(t, string(body), rejected)
	assert.NotContains(t, string(body), current)
	assert.Contains(t, string(body), `"field":"newPassword"`)
}
