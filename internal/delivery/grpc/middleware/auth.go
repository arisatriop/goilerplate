package grpcmiddleware

import (
	"context"
	"strings"

	"goilerplate/config"
	"goilerplate/internal/domain/auth"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/hash"
	jwtService "goilerplate/pkg/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// metadataAuthorization is the conventional gRPC metadata key for a bearer token. gRPC
// lowercases metadata keys, so it is written lowercase here.
const metadataAuthorization = "authorization"

// Auth checks gRPC calls before a handler runs.
//
// In token mode it reuses the HTTP validator and session store, so a call that arrives over
// gRPC ends up with the same user context as the same call over HTTP. Two auth paths that
// disagree about what a token means is how a revoked session keeps working on one of them.
type Auth struct {
	mode           string
	jwtService     *jwtService.JWTService
	sessionService *auth.SessionService
	// secret is the digest of grpc.auth.secret, empty unless the mode is shared_secret.
	secret string
	// exactMethods and wildcardServices hold the public-method allowlist.
	exactMethods     map[string]struct{}
	wildcardServices map[string]struct{}
}

func NewAuth(cfg config.GRPCAuth, jwt *jwtService.JWTService, sessions *auth.SessionService) *Auth {
	secret := ""
	if cfg.ModeOrDefault() == config.GRPCAuthSharedSecret {
		secret = hash.Token(cfg.Secret)
	}

	exact := make(map[string]struct{})
	wildcard := make(map[string]struct{})
	for _, method := range cfg.PublicMethods {
		method = strings.TrimSpace(method)
		if service, found := strings.CutSuffix(method, "/*"); found {
			wildcard[service] = struct{}{}
			continue
		}
		if method != "" {
			exact[method] = struct{}{}
		}
	}

	return &Auth{
		mode:             cfg.ModeOrDefault(),
		jwtService:       jwt,
		sessionService:   sessions,
		secret:           secret,
		exactMethods:     exact,
		wildcardServices: wildcard,
	}
}

// Unary authenticates a unary call and hands the handler an authenticated context.
func (a *Auth) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		authed, err := a.authenticate(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(authed, req)
	}
}

// Stream authenticates a streaming call. A stream is checked once, when it opens: there is no
// later point at which the client can be asked again, so a long-lived stream outlives a
// revocation until it closes.
func (a *Auth) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		authed, err := a.authenticate(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, &authenticatedStream{ServerStream: ss, ctx: authed})
	}
}

// authenticate returns the context the handler should run with, or a gRPC status error.
func (a *Auth) authenticate(ctx context.Context, fullMethod string) (context.Context, error) {
	if a.mode == config.GRPCAuthNone || a.isPublic(fullMethod) {
		return ctx, nil
	}

	if a.mode == config.GRPCAuthSharedSecret {
		if !hash.Equal(metadataValue(ctx, constants.MetadataInternalSecret), a.secret) {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		return ctx, nil
	}

	token, ok := bearerFromMetadata(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	claims, err := a.jwtService.ValidateAccessToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// The same check the HTTP middleware performs, so logging out revokes a login everywhere
	// rather than only on the transport it was revoked from.
	if err := a.sessionService.EnsureActiveForRequest(ctx, claims.SessionID); err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	ctx = context.WithValue(ctx, constants.ContextKeyUserID, claims.UserID)
	ctx = context.WithValue(ctx, constants.ContextKeyUserName, claims.UserName)
	ctx = context.WithValue(ctx, constants.ContextKeySessionID, claims.SessionID)

	return ctx, nil
}

// isPublic reports whether a method is exempt, by exact name or by its service being wildcarded.
func (a *Auth) isPublic(fullMethod string) bool {
	if _, found := a.exactMethods[fullMethod]; found {
		return true
	}
	if index := strings.LastIndex(fullMethod, "/"); index > 0 {
		_, found := a.wildcardServices[fullMethod[:index]]
		return found
	}
	return false
}

// bearerFromMetadata reads the token out of the authorization metadata entry. Every failure
// answers the same way, for the reason given on the HTTP side: the distinctions describe our
// parser, and tell an attacker which guess got further.
func bearerFromMetadata(ctx context.Context) (string, bool) {
	parts := strings.SplitN(metadataValue(ctx, metadataAuthorization), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	return token, token != ""
}

// metadataValue returns the first value for key, or "".
func metadataValue(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if values := md.Get(key); len(values) > 0 {
		return values[0]
	}
	return ""
}

// authenticatedStream replaces a stream's context, which is the only way to hand a streaming
// handler anything the interceptor worked out.
type authenticatedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authenticatedStream) Context() context.Context { return s.ctx }
