// Package refreshtoken decides how the refresh token travels between the API and its client.
//
// In body mode (the default) it is returned in the login and refresh JSON and presented on
// POST /auth/refresh as "Authorization: Bearer <refreshToken>". That suits mobile apps and
// service-to-service clients, which keep secrets in their own storage.
//
// In cookie mode it never appears in a response body. It is set as an HttpOnly cookie scoped to
// /api/v1/auth, so browser JavaScript — and therefore any XSS anywhere on the page — cannot read
// it, and the browser sends it back only to the auth endpoints. The access token stays short-lived
// and in the client's memory.
package refreshtoken

import (
	"net/url"
	"strings"
	"time"

	"goilerplate/pkg/apperr"
	"goilerplate/pkg/constants"

	"github.com/gofiber/fiber/v2"
)

// CookiePath limits the cookie to the auth endpoints: refresh reads it, logout clears it, and no
// other request carries it.
const CookiePath = "/api/v1/auth"

// SameSite values.
const (
	SameSiteStrict = "strict"
	SameSiteLax    = "lax"
	SameSiteNone   = "none"
)

// ErrOriginNotAllowed rejects a cookie-authenticated refresh from a page on another origin.
var ErrOriginNotAllowed = apperr.New(apperr.Forbidden, "origin_not_allowed", constants.MsgForbidden)

// errMissing is what every missing or malformed credential reports, so the answer never says
// which part was wrong.
var errMissing = apperr.New(apperr.Unauthenticated, "unauthorized", constants.MsgUnauthorized)

// Options configures a Transport. The zero value is body mode.
type Options struct {
	// Cookie selects cookie mode.
	Cookie bool
	// Secure marks the cookie Secure and gives it the __Secure- name prefix. Off only for local
	// development over plain http://localhost.
	Secure bool
	// SameSite is strict (the default), lax or none.
	SameSite string
	// AllowedOrigins are the cross-origin pages allowed to refresh with the cookie, normally the
	// CORS allowlist. The API's own origin is always allowed.
	AllowedOrigins []string
}

// Transport reads and writes the refresh token. A nil *Transport is body mode.
type Transport struct {
	opts Options
	name string
}

// New builds a Transport.
func New(opts Options) *Transport {
	opts.SameSite = strings.ToLower(strings.TrimSpace(opts.SameSite))
	if opts.SameSite == "" {
		opts.SameSite = SameSiteStrict
	}

	// The __Secure- prefix makes the browser refuse the cookie unless it arrived over HTTPS with
	// the Secure attribute, so a network attacker cannot plant one over plain HTTP.
	name := "refresh_token"
	if opts.Secure {
		name = "__Secure-refresh_token"
	}
	return &Transport{opts: opts, name: name}
}

// UsesCookie reports whether the refresh token travels in a cookie.
func (t *Transport) UsesCookie() bool { return t != nil && t.opts.Cookie }

// CookieName is the name of the refresh cookie.
func (t *Transport) CookieName() string {
	if t == nil {
		return ""
	}
	return t.name
}

// Read returns the presented refresh token: the cookie in cookie mode, the Bearer header in body
// mode. Each mode accepts only its own transport, so a deployment has one answer to "where does
// the refresh token live".
func (t *Transport) Read(ctx *fiber.Ctx) (string, error) {
	if t.UsesCookie() {
		if token := ctx.Cookies(t.name); token != "" {
			return token, nil
		}
		return "", errMissing
	}
	return BearerToken(ctx)
}

// CheckOrigin guards a cookie-authenticated request against cross-site request forgery.
//
// Browsers always send Origin on a cross-site POST, so a request carrying one must come from the
// API's own origin or an allowed one. A request without Origin is not from a browser page — curl,
// a mobile app, a server — and cannot be a CSRF attempt, so it passes. SameSite already stops a
// cross-site cookie in strict and lax mode; this is the second lock, and the only one under none.
func (t *Transport) CheckOrigin(ctx *fiber.Ctx) error {
	if !t.UsesCookie() {
		return nil
	}

	origin := ctx.Get(fiber.HeaderOrigin)
	if origin == "" || strings.EqualFold(origin, ownOrigin(ctx)) {
		return nil
	}
	for _, allowed := range t.opts.AllowedOrigins {
		if originMatches(allowed, origin) {
			return nil
		}
	}
	return ErrOriginNotAllowed
}

// Issue hands the client its new refresh token: as a cookie in cookie mode. It returns the token
// to put in the response body — the token itself in body mode, "" in cookie mode.
func (t *Transport) Issue(ctx *fiber.Ctx, token string, expiresAt time.Time) string {
	if !t.UsesCookie() {
		return token
	}
	ctx.Cookie(t.cookie(token, expiresAt))
	return ""
}

// Clear removes the refresh cookie after a logout. A no-op in body mode.
func (t *Transport) Clear(ctx *fiber.Ctx) {
	if !t.UsesCookie() {
		return
	}
	ctx.Cookie(t.cookie("", time.Unix(0, 0)))
}

func (t *Transport) cookie(value string, expires time.Time) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     t.name,
		Value:    value,
		Path:     CookiePath,
		Expires:  expires,
		HTTPOnly: true,
		Secure:   t.opts.Secure,
		SameSite: sameSite(t.opts.SameSite),
	}
}

func sameSite(mode string) string {
	switch mode {
	case SameSiteLax:
		return fiber.CookieSameSiteLaxMode
	case SameSiteNone:
		return fiber.CookieSameSiteNoneMode
	default:
		return fiber.CookieSameSiteStrictMode
	}
}

// ownOrigin is scheme://host of the request as the client addressed it. Behind a trusted proxy
// Fiber resolves both from the forwarding headers.
func ownOrigin(ctx *fiber.Ctx) string {
	return ctx.Protocol() + "://" + ctx.Hostname()
}

// originMatches compares an allowlist entry with an Origin header. Entries use Fiber's CORS
// syntax: an exact origin, or https://*.example.com for any subdomain. "*" matches nothing here:
// a wildcard cannot be combined with credentials, so it never authorises a cookie.
func originMatches(allowed, origin string) bool {
	allowed = strings.TrimSpace(allowed)
	if allowed == "" || allowed == "*" {
		return false
	}
	if strings.EqualFold(allowed, origin) {
		return true
	}

	scheme, host, ok := strings.Cut(allowed, "://*.")
	if !ok {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || !strings.EqualFold(parsed.Scheme, scheme) {
		return false
	}
	return strings.HasSuffix(strings.ToLower(parsed.Host), "."+strings.ToLower(host))
}

// BearerToken reads the token out of the Authorization header.
//
// Every failure answers the same way. Telling a caller whether the header was missing, not a
// Bearer scheme, or empty after the scheme describes our parser, not their mistake, and the one
// thing it reliably tells an attacker is which of their guesses got further.
func BearerToken(ctx *fiber.Ctx) (string, error) {
	scheme, token, found := strings.Cut(ctx.Get(fiber.HeaderAuthorization), " ")
	if !found || !strings.EqualFold(scheme, "bearer") {
		return "", errMissing
	}
	if token = strings.TrimSpace(token); token == "" {
		return "", errMissing
	}
	return token, nil
}
