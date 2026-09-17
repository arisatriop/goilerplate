package redact_test

import (
	"testing"

	"goilerplate/pkg/redact"

	"github.com/stretchr/testify/assert"
)

func TestRedactor_Value(t *testing.T) {
	// Arrange
	r := redact.New("pin_code")
	input := map[string]any{
		"email":    "user@example.com",
		"password": "hunter2",
		"data": map[string]any{
			"accessToken":   "access-jwt",
			"Refresh-Token": "refresh-jwt",
			"token_type":    "Bearer",
			"items": []any{
				map[string]any{"pinCode": "1234", "name": "keep"},
			},
		},
	}

	// Act
	got := r.Value(input).(map[string]any)

	// Assert
	assert.Equal(t, "user@example.com", got["email"])
	assert.Equal(t, redact.Mask, got["password"])
	data := got["data"].(map[string]any)
	assert.Equal(t, redact.Mask, data["accessToken"])
	assert.Equal(t, redact.Mask, data["Refresh-Token"])
	assert.Equal(t, "Bearer", data["token_type"])
	item := data["items"].([]any)[0].(map[string]any)
	assert.Equal(t, redact.Mask, item["pinCode"])
	assert.Equal(t, "keep", item["name"])
	assert.Equal(t, "hunter2", input["password"], "input must not be mutated")
}

func TestRedactor_Value_NonMap(t *testing.T) {
	r := redact.New()

	assert.Nil(t, r.Value(nil))
	assert.Equal(t, "plain text", r.Value("plain text"))
}

func TestRedactor_Headers(t *testing.T) {
	// Arrange
	r := redact.New()
	headers := map[string]any{
		"Authorization":     "Bearer abc",
		"x-api-key":         "partner-key",
		"Cookie":            []string{"a=1", "b=2"},
		"X-Internal-Secret": "s3cret",
		"Content-Type":      "application/json",
	}

	// Act
	got := r.Headers(headers)

	// Assert
	assert.Equal(t, redact.Mask, got["Authorization"])
	assert.Equal(t, redact.Mask, got["x-api-key"])
	assert.Equal(t, redact.Mask, got["Cookie"])
	assert.Equal(t, redact.Mask, got["X-Internal-Secret"])
	assert.Equal(t, "application/json", got["Content-Type"])
}

func TestRedactor_URL(t *testing.T) {
	r := redact.New()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"no query", "/api/v1/users", "/api/v1/users"},
		{"safe query", "/users?page=1&limit=10", "/users?page=1&limit=10"},
		{"sensitive query", "/reset?token=abc&email=a%40b.c", "/reset?token=" + redact.Mask + "&email=a%40b.c"},
		{"encoded key", "/x?api%5Fkey=abc", "/x?api%5Fkey=" + redact.Mask},
		{"full url with fragment", "https://h/cb?code=1&secret=x#frag", "https://h/cb?code=1&secret=" + redact.Mask + "#frag"},
		{"key without value", "/x?token", "/x?token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, r.URL(tt.raw))
		})
	}
}

func TestRedactor_EncodedQuery_FormBody(t *testing.T) {
	r := redact.New()

	got := r.EncodedQuery("email=a%40b.c&password=hunter2")

	assert.Equal(t, "email=a%40b.c&password="+redact.Mask, got)
}

func TestRedactor_Query(t *testing.T) {
	r := redact.New()

	got := r.Query(map[string]string{"otp": "123456", "page": "2"})

	assert.Equal(t, redact.Mask, got["otp"])
	assert.Equal(t, "2", got["page"])
}

func TestSetDefault(t *testing.T) {
	original := redact.Default()
	t.Cleanup(func() { redact.SetDefault(original) })

	redact.SetDefault(redact.New("national_id"))
	redact.SetDefault(nil)

	assert.True(t, redact.Default().IsSensitiveField("nationalId"))
}
