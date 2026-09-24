package response_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"goilerplate/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type signupRequest struct {
	Email       string `json:"email" validate:"required,email"`
	NewPassword string `json:"newPassword,omitempty" validate:"required,min=8"`
	Page        int    `query:"page" validate:"gte=1"`
	SessionID   string `params:"id" validate:"uuid"`
	Untagged    string `validate:"required"`
}

func TestFormatValidationErrors_NamesFieldsAsTheClientSentThem(t *testing.T) {
	// Arrange
	req := signupRequest{Email: "not-an-email", NewPassword: "short", SessionID: "x"}

	// Act
	details := response.FormatValidationErrors(response.NewValidator().Struct(&req))

	// Assert
	fields := map[string]string{}
	for _, detail := range details {
		fields[detail.Field] = detail.Tag
	}
	assert.Equal(t, map[string]string{
		"email":       "email",
		"newPassword": "min",
		"page":        "gte",
		"id":          "uuid",
		"Untagged":    "required",
	}, fields)
}

// The submitted value must not come back. For a password field it would put the password in the
// response body, and from there in the response log.
func TestFormatValidationErrors_NeverEchoesTheSubmittedValue(t *testing.T) {
	const password = "hunter2"
	req := signupRequest{Email: "user@example.com", NewPassword: password, Page: 1,
		SessionID: "0190a6f0-0000-7000-8000-000000000001", Untagged: "x"}

	app := fiber.New()
	app.Post("/", func(ctx *fiber.Ctx) error {
		return response.ValidationError(ctx, response.FormatValidationErrors(response.NewValidator().Struct(&req)))
	})
	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/", nil))
	require.NoError(t, err)
	defer resp.Body.Close()

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	encoded, err := json.Marshal(body)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.NotContains(t, string(encoded), password)
	detail := body["errors"].([]any)[0].(map[string]any)
	assert.Equal(t, "newPassword", detail["field"])
	assert.NotContains(t, detail, "value")
	assert.Equal(t, "newPassword must be at least 8 characters", detail["message"])
}

func TestFormatValidationErrors_NonValidationErrorIsAnEmptyList(t *testing.T) {
	details := response.FormatValidationErrors(assert.AnError)

	assert.NotNil(t, details, "an empty list, so errors serialises as [] rather than null")
	assert.Empty(t, details)
}
