package jwt_test

import (
	"testing"
	"time"

	pkgjwt "goilerplate/pkg/jwt"
	"goilerplate/pkg/utils"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	activeAccessSecret  = "active-access-secret-0123456789abcdef"
	activeRefreshSecret = "active-refresh-secret-0123456789abcde"
	oldAccessSecret     = "previous-access-secret-0123456789abc"
	oldRefreshSecret    = "previous-refresh-secret-0123456789ab"

	issuer   = "goilerplate"
	audience = "goilerplate-api"
)

func activeKey() pkgjwt.Key {
	return pkgjwt.Key{ID: "v2", AccessSecret: activeAccessSecret, RefreshSecret: activeRefreshSecret}
}

func previousKey() pkgjwt.Key {
	return pkgjwt.Key{ID: "v1", AccessSecret: oldAccessSecret, RefreshSecret: oldRefreshSecret}
}

func newService(t *testing.T, previous ...pkgjwt.Key) *pkgjwt.JWTService {
	t.Helper()

	svc, err := pkgjwt.NewJWTService(pkgjwt.Config{
		Active:       activeKey(),
		Previous:     previous,
		Issuer:       issuer,
		Audience:     audience,
		AccessExpiry: 15 * time.Minute,
	})
	require.NoError(t, err)
	return svc
}

// signWith mints a token the way a forger would: arbitrary claims, arbitrary key, arbitrary
// algorithm. Every rejection test below differs from a valid token in exactly one way.
func signWith(t *testing.T, method jwtlib.SigningMethod, secret any, kid string, claims *pkgjwt.Claims) string {
	t.Helper()

	token := jwtlib.NewWithClaims(method, claims)
	if kid != "" {
		token.Header["kid"] = kid
	}
	signed, err := token.SignedString(secret)
	require.NoError(t, err)
	return signed
}

func validClaims() *pkgjwt.Claims {
	now := utils.Now()
	return &pkgjwt.Claims{
		UserID:    "user-1",
		SessionID: "session-1",
		Type:      pkgjwt.AccessToken,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    issuer,
			Subject:   "user-1",
			Audience:  jwtlib.ClaimStrings{audience},
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(time.Hour)),
			NotBefore: jwtlib.NewNumericDate(now),
			ID:        utils.GenerateUUID(),
		},
	}
}

func TestGenerateTokenPair_ClaimsAndHeaders(t *testing.T) {
	// Arrange
	svc := newService(t)

	// Act
	pair, err := svc.GenerateTokenPair("user-1", "User One", "user@example.test", "session-1", "device-1", 48*time.Hour)

	// Assert
	require.NoError(t, err)

	access, err := svc.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user-1", access.UserID)
	assert.Equal(t, "session-1", access.SessionID)
	assert.Equal(t, pkgjwt.AccessToken, access.Type)
	assert.Equal(t, issuer, access.Issuer)
	assert.Equal(t, jwtlib.ClaimStrings{audience}, access.Audience)
	assert.Equal(t, pair.AccessTokenID, access.ID, "access token must carry a jti")

	refresh, err := svc.ValidateRefreshToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, pkgjwt.RefreshToken, refresh.Type)
	assert.Equal(t, pair.RefreshTokenID, refresh.ID, "refresh token must carry a jti")
	assert.NotEqual(t, pair.AccessTokenID, pair.RefreshTokenID)

	// The refresh lifetime is the caller's, not the service's
	assert.WithinDuration(t, utils.Now().Add(48*time.Hour), pair.RefreshTokenExpiresAt, time.Minute)
}

func TestValidate_RejectsWrongIssuer(t *testing.T) {
	svc := newService(t)
	claims := validClaims()
	claims.Issuer = "someone-else"

	_, err := svc.ValidateAccessToken(signWith(t, jwtlib.SigningMethodHS256, []byte(activeAccessSecret), "v2", claims))

	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken)
}

func TestValidate_RejectsWrongAudience(t *testing.T) {
	svc := newService(t)
	claims := validClaims()
	claims.Audience = jwtlib.ClaimStrings{"another-api"}

	_, err := svc.ValidateAccessToken(signWith(t, jwtlib.SigningMethodHS256, []byte(activeAccessSecret), "v2", claims))

	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken)
}

func TestValidate_RejectsWrongAlgorithm(t *testing.T) {
	svc := newService(t)

	// "none" is the classic downgrade: no signature at all.
	unsigned := signWith(t, jwtlib.SigningMethodNone, jwtlib.UnsafeAllowNoneSignatureType, "v2", validClaims())
	_, err := svc.ValidateAccessToken(unsigned)
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken, "alg=none must be rejected")

	// A different HMAC size is still the wrong algorithm.
	hs512 := signWith(t, jwtlib.SigningMethodHS512, []byte(activeAccessSecret), "v2", validClaims())
	_, err = svc.ValidateAccessToken(hs512)
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken, "alg=HS512 must be rejected")
}

func TestValidate_RejectsWrongSecret(t *testing.T) {
	svc := newService(t)

	forged := signWith(t, jwtlib.SigningMethodHS256, []byte("attacker-secret-0123456789abcdefgh"), "v2", validClaims())

	_, err := svc.ValidateAccessToken(forged)
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken)
}

func TestValidate_RejectsUnknownOrMissingKeyID(t *testing.T) {
	svc := newService(t)

	unknown := signWith(t, jwtlib.SigningMethodHS256, []byte(activeAccessSecret), "v99", validClaims())
	_, err := svc.ValidateAccessToken(unknown)
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken, "an unknown kid must not fall back to the active key")

	missing := signWith(t, jwtlib.SigningMethodHS256, []byte(activeAccessSecret), "", validClaims())
	_, err = svc.ValidateAccessToken(missing)
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken, "a token without a kid must be rejected")
}

// The refresh secret must not open the access door, and vice versa.
func TestValidate_RejectsSwappedTokenTypes(t *testing.T) {
	svc := newService(t)
	pair, err := svc.GenerateTokenPair("user-1", "User One", "user@example.test", "session-1", "device-1", time.Hour)
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(pair.RefreshToken)
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken, "a refresh token must not authenticate a request")

	_, err = svc.ValidateRefreshToken(pair.AccessToken)
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken, "an access token must not be exchanged for a new pair")
}

func TestValidate_RejectsExpiredToken(t *testing.T) {
	svc := newService(t)
	claims := validClaims()
	past := utils.Now().Add(-2 * time.Hour)
	claims.IssuedAt = jwtlib.NewNumericDate(past)
	claims.NotBefore = jwtlib.NewNumericDate(past)
	claims.ExpiresAt = jwtlib.NewNumericDate(utils.Now().Add(-time.Hour))

	_, err := svc.ValidateAccessToken(signWith(t, jwtlib.SigningMethodHS256, []byte(activeAccessSecret), "v2", claims))

	assert.ErrorIs(t, err, pkgjwt.ErrExpiredToken)
}

// Rotating the active key must not log out everyone holding a token signed with the old one.
func TestValidate_AcceptsTokenSignedWithPreviousKey(t *testing.T) {
	// Arrange: issue a pair while v1 is active
	before, err := pkgjwt.NewJWTService(pkgjwt.Config{
		Active:       previousKey(),
		Issuer:       issuer,
		Audience:     audience,
		AccessExpiry: 15 * time.Minute,
	})
	require.NoError(t, err)

	pair, err := before.GenerateTokenPair("user-1", "User One", "user@example.test", "session-1", "device-1", time.Hour)
	require.NoError(t, err)

	// Act: rotate to v2, keeping v1 for verification
	after := newService(t, previousKey())

	// Assert
	access, err := after.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err, "a token signed with the previous key must still validate")
	assert.Equal(t, "user-1", access.UserID)

	_, err = after.ValidateRefreshToken(pair.RefreshToken)
	assert.NoError(t, err)

	// And once the old key is dropped, its tokens stop working
	assert.NotPanics(t, func() {
		_, err = newService(t).ValidateAccessToken(pair.AccessToken)
	})
	assert.ErrorIs(t, err, pkgjwt.ErrInvalidToken, "dropping a retired key must invalidate its tokens")
}

func TestNewJWTService_RejectsBadConfig(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*pkgjwt.Config)
		wantErr string
	}{
		{"no key id", func(c *pkgjwt.Config) { c.Active.ID = "" }, "active key id is required"},
		{"no access secret", func(c *pkgjwt.Config) { c.Active.AccessSecret = "" }, "needs both an access and a refresh secret"},
		{"no issuer", func(c *pkgjwt.Config) { c.Issuer = "" }, "issuer is required"},
		{"no audience", func(c *pkgjwt.Config) { c.Audience = "" }, "audience is required"},
		{"no access expiry", func(c *pkgjwt.Config) { c.AccessExpiry = 0 }, "access expiry must be greater than 0"},
		{"previous key without id", func(c *pkgjwt.Config) {
			c.Previous = []pkgjwt.Key{{AccessSecret: oldAccessSecret, RefreshSecret: oldRefreshSecret}}
		}, "every previous key needs an id"},
		{"duplicate key id", func(c *pkgjwt.Config) {
			c.Previous = []pkgjwt.Key{{ID: "v2", AccessSecret: oldAccessSecret, RefreshSecret: oldRefreshSecret}}
		}, `duplicate key id "v2"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := pkgjwt.Config{
				Active:       activeKey(),
				Issuer:       issuer,
				Audience:     audience,
				AccessExpiry: 15 * time.Minute,
			}
			tt.mutate(&cfg)

			_, err := pkgjwt.NewJWTService(cfg)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGenerateAccessToken_CarriesJTIAndValidates(t *testing.T) {
	svc := newService(t)

	token, expiresAt, err := svc.GenerateAccessToken("user-1", "User One", "user@example.test", "session-1", "device-1")
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.NotEmpty(t, claims.ID)
	assert.WithinDuration(t, expiresAt, claims.ExpiresAt.Time, time.Second)
}
