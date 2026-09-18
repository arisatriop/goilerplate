package bootstrap

import (
	"path/filepath"
	"testing"

	"goilerplate/config"
	grpcmiddleware "goilerplate/internal/delivery/grpc/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// middlewareCount counts the handlers registered with app.Use, as seen by GET requests.
func middlewareCount(t *testing.T, cfg *config.Config) int {
	t.Helper()

	app := NewFiber(cfg)
	stack := app.Stack()
	count := 0
	for i, method := range app.Config().RequestMethods {
		if method != fiber.MethodGet {
			continue
		}
		for _, route := range stack[i] {
			if route.Path == "/" {
				count += len(route.Handlers)
			}
		}
	}
	return count
}

func TestNewFiber_OTelMiddlewareOnlyWhenEnabled(t *testing.T) {
	disabled := middlewareCount(t, &config.Config{})
	enabled := middlewareCount(t, &config.Config{OTel: config.OTel{Enabled: true}})

	assert.Positive(t, disabled, "recover middleware is always registered")
	assert.Equal(t, disabled+1, enabled, "otel.enabled adds exactly the otelfiber middleware")
}

func TestNewGrpcServer_WithAndWithoutOTel(t *testing.T) {
	for _, otelEnabled := range []bool{false, true} {
		cfg := &config.Config{OTel: config.OTel{Enabled: otelEnabled}}
		server, err := NewGrpcServer(cfg, grpcmiddleware.NewAuth(cfg.GRPC.Auth, nil, nil))
		require.NoError(t, err)
		assert.NotNil(t, server)
		server.Stop()
	}
}

// Reflection publishes the whole service surface to anyone who can reach the port, so it is off
// unless asked for. It used to key off app.env, which answers a different question.
func TestNewGrpcServer_ReflectionOffUnlessConfigured(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cfg      *config.Config
		wantRefl bool
	}{
		{"default", &config.Config{}, false},
		{"non-production env alone does not enable it", &config.Config{App: config.App{Env: "local"}}, false},
		{"explicitly enabled", &config.Config{GRPC: config.GRPC{Reflection: true}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server, err := NewGrpcServer(tc.cfg, grpcmiddleware.NewAuth(tc.cfg.GRPC.Auth, nil, nil))
			require.NoError(t, err)
			defer server.Stop()

			_, registered := server.GetServiceInfo()["grpc.reflection.v1.ServerReflection"]
			assert.Equal(t, tc.wantRefl, registered)
		})
	}
}

// A named certificate that cannot be read must stop startup rather than quietly serve plaintext
// on a port the operator believes is encrypted.
func TestNewGrpcServer_MissingTLSCertificateIsAnError(t *testing.T) {
	cfg := &config.Config{GRPC: config.GRPC{TLS: config.GRPCTLS{
		Enabled:  true,
		CertFile: filepath.Join(t.TempDir(), "absent.pem"),
		KeyFile:  filepath.Join(t.TempDir(), "absent.key"),
	}}}

	server, err := NewGrpcServer(cfg, grpcmiddleware.NewAuth(cfg.GRPC.Auth, nil, nil))

	require.Error(t, err)
	assert.Nil(t, server)
}

func TestPartnerRoutesEnabled(t *testing.T) {
	assert.False(t, PartnerRoutesEnabled(&config.Config{}))
	assert.False(t, PartnerRoutesEnabled(&config.Config{Apikeys: map[string]string{}}))
	assert.True(t, PartnerRoutesEnabled(&config.Config{Apikeys: map[string]string{"partner1": "key"}}))
}
