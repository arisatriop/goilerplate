package grpcmiddleware

import (
	"context"
	"testing"
	"time"

	"goilerplate/config"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/constants"
	jwtpkg "goilerplate/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	testMethod       = "/bar.v1.BarService/List"
	testHealthMethod = "/grpc.health.v1.Health/Check"
	testGRPCSecret   = "grpc-shared-secret-8d1f4b6a2c9e7053"
)

// sessionRepo answers the one repository call the session check makes.
type sessionRepo struct {
	auth.Repository
	session *auth.UserSession
}

func (r *sessionRepo) GetSessionByID(context.Context, string) (*auth.UserSession, error) {
	return r.session, nil
}

// noopStore keeps the session check going to the repository, which is where revocation lives.
type noopStore struct{}

func (noopStore) Get(context.Context, string) (*auth.UserSession, bool, error) {
	return nil, false, nil
}
func (noopStore) Set(context.Context, *auth.UserSession) error { return nil }
func (noopStore) Delete(context.Context, string) error         { return nil }
func (noopStore) DeleteByUser(context.Context, string) error   { return nil }

func testJWT(t *testing.T) *jwtpkg.JWTService {
	t.Helper()

	service, err := jwtpkg.NewJWTService(jwtpkg.Config{
		Active: jwtpkg.Key{
			ID:            "v1",
			AccessSecret:  "k3Jv9QpX2mZr7TbW5nLc8HsYd4FgA1uE",
			RefreshSecret: "Pz6Rw1YtN8qLm3XcV7bK2jHf5Gd9Sa4U",
		},
		Issuer:       "goilerplate",
		Audience:     "goilerplate-api",
		AccessExpiry: 15 * time.Minute,
	})
	require.NoError(t, err)

	return service
}

// activeSession is what the repository returns for a login that has not been revoked.
func activeSession(id string) *auth.UserSession {
	return &auth.UserSession{ID: id, UserID: "u1", IsActive: true, ExpiresAt: time.Now().Add(time.Hour)}
}

// callUnary runs one call through the interceptor and reports the context the handler saw.
func callUnary(t *testing.T, interceptor *Auth, method string, md metadata.MD) (context.Context, error) {
	t.Helper()

	ctx := context.Background()
	if md != nil {
		ctx = metadata.NewIncomingContext(ctx, md)
	}

	var handlerCtx context.Context
	_, err := interceptor.Unary()(ctx, nil, &grpc.UnaryServerInfo{FullMethod: method},
		func(ctx context.Context, _ any) (any, error) {
			handlerCtx = ctx
			return "ok", nil
		})

	return handlerCtx, err
}

func assertUnauthenticated(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

// The default is the strict mode: a port that answers before it has been told what to require
// should refuse, not serve.
func TestNewAuth_DefaultsToTokenMode(t *testing.T) {
	interceptor := NewAuth(config.GRPCAuth{}, testJWT(t), nil)

	_, err := callUnary(t, interceptor, testMethod, nil)

	assertUnauthenticated(t, err)
}

func TestAuth_TokenMode(t *testing.T) {
	jwt := testJWT(t)
	sessions := auth.NewSessionService(&sessionRepo{session: activeSession("s1")}, noopStore{}, true)
	interceptor := NewAuth(config.GRPCAuth{Mode: config.GRPCAuthToken}, jwt, sessions)

	token, _, err := jwt.GenerateAccessToken("u1", "User", "u@example.test", "s1", "d1")
	require.NoError(t, err)

	t.Run("a valid token authenticates and reaches the handler", func(t *testing.T) {
		handlerCtx, err := callUnary(t, interceptor, testMethod,
			metadata.Pairs(metadataAuthorization, "Bearer "+token))

		require.NoError(t, err)
		require.NotNil(t, handlerCtx)

		// The same context keys the HTTP middleware sets, so a handler cannot tell which
		// transport it was reached over.
		assert.Equal(t, "u1", handlerCtx.Value(constants.ContextKeyUserID))
		assert.Equal(t, "User", handlerCtx.Value(constants.ContextKeyUserName))
		assert.Equal(t, "s1", handlerCtx.Value(constants.ContextKeySessionID))
	})

	t.Run("rejections", func(t *testing.T) {
		for name, md := range map[string]metadata.MD{
			"no metadata at all":     nil,
			"no authorization entry": metadata.Pairs("x-other", "value"),
			"empty authorization":    metadata.Pairs(metadataAuthorization, ""),
			"bearer with no token":   metadata.Pairs(metadataAuthorization, "Bearer "),
			"token without scheme":   metadata.Pairs(metadataAuthorization, token),
			"wrong scheme":           metadata.Pairs(metadataAuthorization, "Basic "+token),
			"tampered token":         metadata.Pairs(metadataAuthorization, "Bearer "+token+"x"),
			"not a token":            metadata.Pairs(metadataAuthorization, "Bearer not-a-jwt"),
		} {
			t.Run(name, func(t *testing.T) {
				_, err := callUnary(t, interceptor, testMethod, md)
				assertUnauthenticated(t, err)
			})
		}
	})

	t.Run("the scheme is case-insensitive, as RFC 7235 requires", func(t *testing.T) {
		_, err := callUnary(t, interceptor, testMethod,
			metadata.Pairs(metadataAuthorization, "bearer "+token))

		assert.NoError(t, err)
	})
}

// The point of reusing the HTTP session store: logging out has to revoke a login on every
// transport, not only the one it was revoked from.
func TestAuth_TokenMode_RevokedSessionIsRejected(t *testing.T) {
	jwt := testJWT(t)
	revoked := activeSession("s1")
	revoked.IsActive = false

	sessions := auth.NewSessionService(&sessionRepo{session: revoked}, noopStore{}, true)
	interceptor := NewAuth(config.GRPCAuth{Mode: config.GRPCAuthToken}, jwt, sessions)

	token, _, err := jwt.GenerateAccessToken("u1", "User", "u@example.test", "s1", "d1")
	require.NoError(t, err)

	_, err = callUnary(t, interceptor, testMethod,
		metadata.Pairs(metadataAuthorization, "Bearer "+token))

	assertUnauthenticated(t, err)
}

// A refresh token is not an access token. The parser pins the token type, so this is rejected
// before any later check has a chance to be forgotten.
func TestAuth_TokenMode_RefreshTokenIsRejected(t *testing.T) {
	jwt := testJWT(t)
	sessions := auth.NewSessionService(&sessionRepo{session: activeSession("s1")}, noopStore{}, true)
	interceptor := NewAuth(config.GRPCAuth{Mode: config.GRPCAuthToken}, jwt, sessions)

	pair, err := jwt.GenerateTokenPair("u1", "User", "u@example.test", "s1", "d1", time.Hour)
	require.NoError(t, err)

	_, err = callUnary(t, interceptor, testMethod,
		metadata.Pairs(metadataAuthorization, "Bearer "+pair.RefreshToken))

	assertUnauthenticated(t, err)
}

func TestAuth_SharedSecretMode(t *testing.T) {
	interceptor := NewAuth(config.GRPCAuth{
		Mode:   config.GRPCAuthSharedSecret,
		Secret: testGRPCSecret,
	}, nil, nil)

	tests := map[string]struct {
		md   metadata.MD
		want codes.Code
	}{
		"correct secret":      {metadata.Pairs(constants.MetadataInternalSecret, testGRPCSecret), codes.OK},
		"no metadata":         {nil, codes.Unauthenticated},
		"wrong secret":        {metadata.Pairs(constants.MetadataInternalSecret, "wrong"), codes.Unauthenticated},
		"one character extra": {metadata.Pairs(constants.MetadataInternalSecret, testGRPCSecret+"x"), codes.Unauthenticated},
		"one character short": {metadata.Pairs(constants.MetadataInternalSecret, testGRPCSecret[:len(testGRPCSecret)-1]), codes.Unauthenticated},
		"a token instead":     {metadata.Pairs(metadataAuthorization, "Bearer "+testGRPCSecret), codes.Unauthenticated},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := callUnary(t, interceptor, testMethod, tc.md)
			assert.Equal(t, tc.want, status.Code(err))
		})
	}
}

func TestAuth_NoneModeChecksNothing(t *testing.T) {
	interceptor := NewAuth(config.GRPCAuth{Mode: config.GRPCAuthNone}, nil, nil)

	_, err := callUnary(t, interceptor, testMethod, nil)

	assert.NoError(t, err)
}

func TestAuth_PublicMethodAllowlist(t *testing.T) {
	interceptor := NewAuth(config.GRPCAuth{
		Mode: config.GRPCAuthToken,
		PublicMethods: []string{
			"/hello.v1.HelloService/SayHello",
			"/grpc.health.v1.Health/*",
		},
	}, testJWT(t), nil)

	tests := map[string]struct {
		method string
		want   codes.Code
	}{
		"exact match is exempt":                     {"/hello.v1.HelloService/SayHello", codes.OK},
		"wildcard covers every method of a service": {testHealthMethod, codes.OK},
		"wildcard covers another method of it":      {"/grpc.health.v1.Health/Watch", codes.OK},
		"a sibling method is not exempt":            {"/hello.v1.HelloService/SayGoodbye", codes.Unauthenticated},
		"a different service is not exempt":         {testMethod, codes.Unauthenticated},
		"a service that merely shares a prefix":     {"/grpc.health.v1.HealthAdmin/Check", codes.Unauthenticated},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := callUnary(t, interceptor, tc.method, nil)
			assert.Equal(t, tc.want, status.Code(err))
		})
	}
}

// fakeStream is the minimum a streaming interceptor needs to run.
type fakeStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *fakeStream) Context() context.Context { return s.ctx }

// Streams are authenticated too. An interceptor that only covered unary calls would leave every
// streaming method open while looking like it protected the server.
func TestAuth_StreamIsAuthenticatedAndCarriesTheContext(t *testing.T) {
	jwt := testJWT(t)
	sessions := auth.NewSessionService(&sessionRepo{session: activeSession("s1")}, noopStore{}, true)
	interceptor := NewAuth(config.GRPCAuth{Mode: config.GRPCAuthToken}, jwt, sessions)

	token, _, err := jwt.GenerateAccessToken("u1", "User", "u@example.test", "s1", "d1")
	require.NoError(t, err)

	run := func(md metadata.MD) (context.Context, error) {
		ctx := context.Background()
		if md != nil {
			ctx = metadata.NewIncomingContext(ctx, md)
		}

		var streamCtx context.Context
		err := interceptor.Stream()(nil, &fakeStream{ctx: ctx},
			&grpc.StreamServerInfo{FullMethod: testMethod},
			func(_ any, stream grpc.ServerStream) error {
				streamCtx = stream.Context()
				return nil
			})

		return streamCtx, err
	}

	t.Run("rejected without a token", func(t *testing.T) {
		_, err := run(nil)
		assertUnauthenticated(t, err)
	})

	t.Run("the handler's stream carries the authenticated context", func(t *testing.T) {
		streamCtx, err := run(metadata.Pairs(metadataAuthorization, "Bearer "+token))

		require.NoError(t, err)
		require.NotNil(t, streamCtx)
		assert.Equal(t, "u1", streamCtx.Value(constants.ContextKeyUserID))
		assert.Equal(t, "s1", streamCtx.Value(constants.ContextKeySessionID))
	})
}
