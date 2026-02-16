package constants

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

// Context Keys
const (
	ContextKeyRequestID ContextKey = "request_id"
	ContextKeyUserID    ContextKey = "user_id"
	ContextKeyUserName  ContextKey = "user_name"
	ContextTokenHash    ContextKey = "token_hash"
	ContextKeySessionID ContextKey = "session_id"
	ContextKeyStoreID   ContextKey = "store_id"
)

const (
	HeaderRequestID = "X-Request-Id"
)
