package router

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"goilerplate/config"
	"goilerplate/internal/bootstrap"
	database "goilerplate/internal/bootstrap/database"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// unreachableDB returns a GORM handle pointed at a port nobody is listening on. Automatic
// pinging is disabled so gorm.Open succeeds and the failure happens where the readiness check
// makes it happen — with a driver error that names the host, port, user and database.
func unreachableDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := database.PostgresDSN(config.DB{
		Host:     "127.0.0.1",
		Port:     1,
		Username: "secret-user",
		Password: "secret-password",
		Name:     "secret-database",
		SSLMode:  "disable",
	})

	gdb, err := gorm.Open(gormPostgres.Open(dsn), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	return gdb
}

func newProbeApp(t *testing.T, gdb *gorm.DB) *fiber.App {
	t.Helper()

	registry := &RouteRegistry{App: &bootstrap.App{
		DB:     &database.DB{GDB: gdb},
		Config: &config.Config{App: config.App{Name: "goilerplate", Version: "test"}},
	}}

	app := fiber.New()
	app.Get("/livez", registry.live)
	app.Get("/readyz", registry.ready)
	return app
}

func probe(t *testing.T, app *fiber.App, path string) (int, string) {
	t.Helper()

	res, err := app.Test(httptest.NewRequest(fiber.MethodGet, path, nil), 15000)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return res.StatusCode, string(body)
}

// Liveness must not depend on anything outside the process: restarting this pod does not repair
// a database, and a probe that says otherwise turns one outage into a restart loop.
func TestLive_IsUpEvenWhenTheDatabaseIsUnreachable(t *testing.T) {
	// Arrange
	app := newProbeApp(t, unreachableDB(t))

	// Act
	status, body := probe(t, app, "/livez")

	// Assert
	assert.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, `"status":"ok"`)
}

func TestReady_ReturnsServiceUnavailableWhenADependencyIsDown(t *testing.T) {
	// Arrange
	app := newProbeApp(t, unreachableDB(t))

	// Act
	status, body := probe(t, app, "/readyz")

	// Assert: the status code carries the verdict, so a probe need not parse the body.
	assert.Equal(t, fiber.StatusServiceUnavailable, status)

	var payload struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	assert.Equal(t, "unavailable", payload.Status)
	assert.Equal(t, "unhealthy", payload.Checks["postgresql"])
}

// The endpoint is unauthenticated. It used to embed err.Error() from the driver, which names
// the host, port, database and user — free reconnaissance for anyone who can reach the port.
func TestReady_LeaksNothingAboutWhyItIsUnhealthy(t *testing.T) {
	// Arrange
	app := newProbeApp(t, unreachableDB(t))

	// Act
	status, body := probe(t, app, "/readyz")

	// Assert
	require.Equal(t, fiber.StatusServiceUnavailable, status)
	for _, secret := range []string{"secret-user", "secret-password", "secret-database", "127.0.0.1", "connection refused"} {
		assert.NotContains(t, body, secret, "the response must not describe the failure")
	}
}

func TestReady_IsReadyWhenEveryConfiguredDependencyAnswers(t *testing.T) {
	// Arrange: no database and no Redis configured, so there is nothing that can be down.
	app := newProbeApp(t, nil)

	// Act
	status, body := probe(t, app, "/readyz")

	// Assert
	assert.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, `"status":"ok"`)
}
