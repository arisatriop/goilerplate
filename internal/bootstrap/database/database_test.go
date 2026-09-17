package bootstrap

import (
	"testing"

	"goilerplate/config"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresDSN_ParsesSpecialCharacters(t *testing.T) {
	// Arrange
	db := config.DB{
		Host:     "db.internal",
		Port:     6543,
		Name:     "app db",
		Username: "app_user",
		Password: `p@ss w'rd\with=sign`,
		SSLMode:  "disable",
	}

	// Act
	parsed, err := pgconn.ParseConfig(PostgresDSN(db))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "db.internal", parsed.Host)
	assert.Equal(t, uint16(6543), parsed.Port)
	assert.Equal(t, "app db", parsed.Database)
	assert.Equal(t, "app_user", parsed.User)
	assert.Equal(t, `p@ss w'rd\with=sign`, parsed.Password)
	assert.Nil(t, parsed.TLSConfig, "sslmode=disable turns TLS off")
	assert.Equal(t, "UTC", parsed.RuntimeParams["timezone"], "session time zone is always UTC")
	assert.Contains(t, PostgresDSN(db), " timezone=UTC", "gorm.io/driver/postgres rejects a quoted time zone")
}

func TestPostgresDSN_UsesConfiguredSSLMode(t *testing.T) {
	db := config.DB{Host: "localhost", Port: 5432, Name: "app", Username: "app", SSLMode: "require"}

	parsed, err := pgconn.ParseConfig(PostgresDSN(db))

	require.NoError(t, err)
	require.NotNil(t, parsed.TLSConfig, "sslmode=require enables TLS")
	assert.Contains(t, PostgresDSN(db), "sslmode='require'")
}

func TestPostgresDSN_DefaultSSLMode(t *testing.T) {
	assert.Contains(t, PostgresDSN(config.DB{Host: "localhost", Port: 5432}), "sslmode='prefer'")
}
