// Package redact masks secrets (passwords, tokens, API keys) in values before they are logged.
package redact

import (
	"net/url"
	"strings"
	"sync/atomic"
)

// Mask replaces every redacted value.
const Mask = "[REDACTED]"

// Omitted replaces a request or response body that is not logged at all.
const Omitted = "[OMITTED]"

// DefaultFields are always redacted, in addition to fields passed to New.
var DefaultFields = []string{
	"password",
	"current_password",
	"new_password",
	"old_password",
	"confirm_password",
	"password_confirmation",
	"access_token",
	"refresh_token",
	"id_token",
	"token",
	"otp",
	"secret",
	"client_secret",
	"api_key",
	"private_key",
}

// DefaultHeaders are always redacted.
var DefaultHeaders = []string{
	"Authorization",
	"Proxy-Authorization",
	"Cookie",
	"Set-Cookie",
	"X-Api-Key",
	"X-Internal-Secret",
}

// Redactor masks sensitive fields, headers, and query parameters.
// Names are matched case-insensitively, ignoring '_' and '-', so "refresh_token",
// "refreshToken" (protojson), and "Refresh-Token" all match.
type Redactor struct {
	fields  map[string]struct{}
	headers map[string]struct{}
}

var defaultRedactor atomic.Pointer[Redactor]

func init() {
	defaultRedactor.Store(New())
}

// New creates a Redactor with the default fields plus extraFields.
func New(extraFields ...string) *Redactor {
	r := &Redactor{
		fields:  make(map[string]struct{}),
		headers: make(map[string]struct{}),
	}
	for _, name := range DefaultFields {
		r.fields[normalize(name)] = struct{}{}
	}
	for _, name := range extraFields {
		if name = strings.TrimSpace(name); name != "" {
			r.fields[normalize(name)] = struct{}{}
		}
	}
	for _, name := range DefaultHeaders {
		r.headers[normalize(name)] = struct{}{}
	}
	return r
}

// Default returns the process-wide Redactor used by the loggers.
func Default() *Redactor {
	return defaultRedactor.Load()
}

// SetDefault replaces the process-wide Redactor. Call it once at startup.
func SetDefault(r *Redactor) {
	if r != nil {
		defaultRedactor.Store(r)
	}
}

// IsSensitiveField reports whether a field or parameter with this name must be redacted.
func (r *Redactor) IsSensitiveField(name string) bool {
	_, ok := r.fields[normalize(name)]
	return ok
}

// IsSensitiveHeader reports whether a header with this name must be redacted.
func (r *Redactor) IsSensitiveHeader(name string) bool {
	key := normalize(name)
	if _, ok := r.headers[key]; ok {
		return true
	}
	_, ok := r.fields[key]
	return ok
}

// Value returns a copy of v with sensitive map keys masked at any depth.
// It understands the shapes produced by encoding/json: map[string]any and []any.
func (r *Redactor) Value(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, val := range typed {
			if r.IsSensitiveField(key) {
				out[key] = Mask
				continue
			}
			out[key] = r.Value(val)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, val := range typed {
			out[i] = r.Value(val)
		}
		return out
	default:
		return v
	}
}

// Headers returns a copy of headers with sensitive header values masked.
func (r *Redactor) Headers(headers map[string]any) map[string]any {
	out := make(map[string]any, len(headers))
	for key, val := range headers {
		if r.IsSensitiveHeader(key) {
			out[key] = Mask
			continue
		}
		out[key] = val
	}
	return out
}

// Query returns a copy of params with sensitive parameter values masked.
func (r *Redactor) Query(params map[string]string) map[string]string {
	out := make(map[string]string, len(params))
	for key, val := range params {
		if r.IsSensitiveField(key) {
			out[key] = Mask
			continue
		}
		out[key] = val
	}
	return out
}

// URL masks sensitive query parameter values in a URL or request URI.
func (r *Redactor) URL(raw string) string {
	idx := strings.IndexByte(raw, '?')
	if idx < 0 {
		return raw
	}

	rest := raw[idx+1:]
	fragment := ""
	if hash := strings.IndexByte(rest, '#'); hash >= 0 {
		rest, fragment = rest[:hash], rest[hash:]
	}

	return raw[:idx+1] + r.EncodedQuery(rest) + fragment
}

// EncodedQuery masks sensitive values in an encoded query string or
// application/x-www-form-urlencoded body, preserving parameter order.
func (r *Redactor) EncodedQuery(query string) string {
	if query == "" {
		return query
	}

	pairs := strings.Split(query, "&")
	for i, pair := range pairs {
		key, _, hasValue := strings.Cut(pair, "=")
		name, err := url.QueryUnescape(key)
		if err != nil {
			name = key
		}
		if hasValue && r.IsSensitiveField(name) {
			pairs[i] = key + "=" + Mask
		}
	}
	return strings.Join(pairs, "&")
}

var separatorRemover = strings.NewReplacer("_", "", "-", "")

func normalize(name string) string {
	return separatorRemover.Replace(strings.ToLower(name))
}
