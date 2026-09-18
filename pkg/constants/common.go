package constants

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Context Keys
const (
	ContextKeyRequestID ContextKey = "request_id"
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyUserName  ContextKey = "user_name"
	ContextKeySessionID ContextKey = "session_id"
	ContextKeyStoreID   ContextKey = "store_id"

	// ContextKeyClientIP and ContextKeyUserAgent carry who the request came from, so code far
	// from the transport — the auth domain, for instance — can record it in a security event
	// without taking a *fiber.Ctx or re-parsing headers it cannot safely interpret.
	ContextKeyClientIP  ContextKey = "client_ip"
	ContextKeyUserAgent ContextKey = "user_agent"
)

const (
	HeaderRequestID   = "X-Request-Id"
	HeaderAPIKey      = "x-api-key"
	HeaderServiceName = "X-Service-Name"
	// HeaderInternalSecret is the shared secret for /internal routes when
	// internal_auth.mode=shared_secret. Already redacted from logs (see pkg/redact).
	HeaderInternalSecret = "X-Internal-Secret"
)
