package middleware

import (
	"net/http/httptest"
	"testing"

	"goilerplate/config"

	"goilerplate/pkg/apikey"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/hash"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPartnerName = "acme"
	testPartnerKey  = "partner-key-3f9a2c7e1b4d8a6f0c5e"
)

type partnerCall struct {
	status    int
	userID    string
	userName  string
	localID   string
	localName string
}

// callPartnerRoute runs one request through PartnerAuthenticate and reports what the handler
// behind it would see.
func callPartnerRoute(t *testing.T, keys map[string]string, header string) partnerCall {
	t.Helper()

	auth := NewAuth(nil, nil, nil, nil, keys, config.InternalAuth{})

	var got partnerCall
	app := fiber.New()
	app.Get("/partner", auth.PartnerAuthenticate(), func(ctx *fiber.Ctx) error {
		userCtx := ctx.UserContext()
		got.userID, _ = userCtx.Value(constants.ContextKeyUserID).(string)
		got.userName, _ = userCtx.Value(constants.ContextKeyUserName).(string)
		got.localID, _ = ctx.Locals(string(constants.ContextKeyUserID)).(string)
		got.localName, _ = ctx.Locals(string(constants.ContextKeyUserName)).(string)
		return ctx.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/partner", nil)
	if header != "" {
		req.Header.Set(constants.HeaderAPIKey, header)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	got.status = resp.StatusCode

	return got
}

// The point of the change: the key authenticates the request and then stops existing. It used
// to be stored as the user ID, which meant a live credential was written into every log line
// the request produced — logger.baseAttrs puts user_id on all of them — and into any cache or
// limiter keyed on the caller.
func TestPartnerAuthenticate_PutsTheNameInContextNotTheKey(t *testing.T) {
	got := callPartnerRoute(t, map[string]string{testPartnerName: testPartnerKey}, testPartnerKey)

	require.Equal(t, fiber.StatusOK, got.status)

	assert.Equal(t, testPartnerName, got.userID)
	assert.Equal(t, testPartnerName, got.userName)
	assert.Equal(t, testPartnerName, got.localID)
	assert.Equal(t, testPartnerName, got.localName)

	for field, value := range map[string]string{
		"user context user_id":   got.userID,
		"user context user_name": got.userName,
		"locals user_id":         got.localID,
		"locals user_name":       got.localName,
	} {
		assert.NotContains(t, value, testPartnerKey, "%s still carries the raw API key", field)
	}
}

func TestPartnerAuthenticate_RejectsMissingAndWrongKeys(t *testing.T) {
	keys := map[string]string{testPartnerName: testPartnerKey}

	for _, header := range []string{
		"",
		"wrong-key",
		testPartnerKey + "x",
		testPartnerKey[:len(testPartnerKey)-1],
	} {
		got := callPartnerRoute(t, keys, header)
		assert.Equal(t, fiber.StatusUnauthorized, got.status, "must reject %q", header)
	}
}

// A deployment can configure the digest instead of the key, so the config file never holds a
// usable credential.
func TestPartnerAuthenticate_AcceptsAPreHashedConfiguredKey(t *testing.T) {
	keys := map[string]string{testPartnerName: apikey.HashPrefix + hash.Token(testPartnerKey)}

	got := callPartnerRoute(t, keys, testPartnerKey)

	assert.Equal(t, fiber.StatusOK, got.status)
	assert.Equal(t, testPartnerName, got.userID)
}

// With no partners configured the route is closed, including to a request that sends no header
// at all.
func TestPartnerAuthenticate_NoConfiguredPartnersRejectsEveryone(t *testing.T) {
	assert.Equal(t, fiber.StatusUnauthorized, callPartnerRoute(t, nil, "").status)
	assert.Equal(t, fiber.StatusUnauthorized, callPartnerRoute(t, nil, testPartnerKey).status)
}
