package middleware

import (
	"net/http/httptest"
	"testing"

	"goilerplate/config"
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testInternalSecret = "internal-shared-secret-6b2f8e4d1a97c530"

type internalCall struct {
	status int
	caller string
}

// callInternalRoute runs one request through InternalAuthenticate and reports what reached the
// handler behind it.
func callInternalRoute(t *testing.T, internalAuth config.InternalAuth, headers map[string]string) internalCall {
	t.Helper()

	auth := NewAuth(nil, nil, nil, nil, nil, internalAuth)

	var got internalCall
	app := fiber.New()
	app.Get("/internal", auth.InternalAuthenticate(), func(ctx *fiber.Ctx) error {
		got.caller, _ = ctx.UserContext().Value(constants.ContextKeyUserName).(string)
		return ctx.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/internal", nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	got.status = resp.StatusCode

	return got
}

// The default stays open, because D5 puts /internal behind the gateway's allowlist and a service
// whose in-cluster callers suddenly need a secret they were never given is worse than the status
// quo. Warnings(), not this middleware, is what says so in production.
func TestInternalAuthenticate_NoneModeAllowsEveryone(t *testing.T) {
	got := callInternalRoute(t, config.InternalAuth{}, nil)

	assert.Equal(t, fiber.StatusOK, got.status)
	assert.Equal(t, "system", got.caller)
}

func TestInternalAuthenticate_SharedSecretRequiresTheHeader(t *testing.T) {
	mode := config.InternalAuth{Mode: config.InternalAuthSharedSecret, Secret: testInternalSecret}

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"correct secret", testInternalSecret, fiber.StatusOK},
		{"no header", "", fiber.StatusUnauthorized},
		{"wrong secret", "wrong", fiber.StatusUnauthorized},
		{"secret with a character appended", testInternalSecret + "x", fiber.StatusUnauthorized},
		{"secret one character short", testInternalSecret[:len(testInternalSecret)-1], fiber.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{}
			if tc.header != "" {
				headers[constants.HeaderInternalSecret] = tc.header
			}

			assert.Equal(t, tc.want, callInternalRoute(t, mode, headers).status)
		})
	}
}

// A mode the config does not recognise is rejected at startup, so it cannot reach here. What
// must not happen is the middleware treating an unknown mode as "no secret required".
func TestInternalAuthenticate_UnknownModeDoesNotRequireASecret(t *testing.T) {
	got := callInternalRoute(t, config.InternalAuth{Mode: "mtls", Secret: testInternalSecret}, nil)

	assert.Equal(t, fiber.StatusOK, got.status, "config validation is what rejects this mode")
}

func TestInternalAuthenticate_RecordsTheCallingService(t *testing.T) {
	got := callInternalRoute(t, config.InternalAuth{}, map[string]string{
		constants.HeaderServiceName: "billing-worker",
	})

	require.Equal(t, fiber.StatusOK, got.status)
	assert.Equal(t, "billing-worker", got.caller)
}

// X-Service-Name is an unverified claim that lands in every log line the request produces, so it
// is bounded and stripped rather than taken as given.
func TestInternalCallerName_SanitisesTheHeader(t *testing.T) {
	long := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 60 chars

	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"empty falls back", "", defaultInternalCaller},
		{"whitespace only falls back", "   ", defaultInternalCaller},
		{"plain name", "billing-worker", "billing-worker"},
		{"dots and underscores kept", "billing_worker.v2", "billing_worker.v2"},
		{"surrounding space trimmed", "  billing-worker  ", "billing-worker"},
		{"newlines stripped", "billing\nworker", "billingworker"},
		{"quotes and braces stripped", `bill"ing{}`, "billing"},
		{"nothing usable falls back", `{"":}`, defaultInternalCaller},
		{"over-long name truncated", long, long[:maxServiceNameLength]},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, internalCallerName(tc.header))
		})
	}
}
