package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadExample decodes an example file strictly, so unknown or misspelled keys fail the test.
func loadExample(t *testing.T, path string) *Config {
	t.Helper()

	v := viper.New()
	v.SetConfigFile(path)
	require.NoError(t, v.ReadInConfig())

	var cfg Config
	require.NoError(t, v.UnmarshalExact(&cfg))
	return &cfg
}

func TestConfigExample_MinimalIsValid(t *testing.T) {
	// Arrange
	cfg := loadExample(t, "config.example.yaml")

	// Act
	err := cfg.Validate()

	// Assert
	assert.NoError(t, err, "copying config.example.yaml plus a database must be enough to run")
	assert.False(t, cfg.Redis.Enabled)
	assert.False(t, cfg.GRPC.Enabled)
	assert.False(t, cfg.OTel.Enabled)
	assert.Equal(t, "local", cfg.FileSystem.Driver)
	assert.Empty(t, cfg.Services)
}

func TestConfigExample_MinimalSecretsRejectedInProduction(t *testing.T) {
	cfg := loadExample(t, "config.example.yaml")
	cfg.App.Env = "production"

	err := cfg.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt.access_secret looks like an example value")
	assert.Contains(t, err.Error(), "jwt.refresh_secret looks like an example value")
}

func TestConfigExample_FullDocumentsEveryOption(t *testing.T) {
	v := viper.New()
	v.SetConfigFile("config.full.example.yaml")
	require.NoError(t, v.ReadInConfig())

	for _, key := range configKeys(reflect.TypeOf(Config{}), "") {
		assert.True(t, v.IsSet(key), "config.full.example.yaml is missing %q", key)
	}

	cfg := loadExample(t, "config.full.example.yaml")

	assert.True(t, cfg.Redis.Enabled)
	assert.True(t, cfg.GRPC.Enabled)
	assert.True(t, cfg.OTel.Enabled)
	assert.NotEmpty(t, cfg.Server.CORS.AllowOrigin)
	assert.NotEmpty(t, cfg.Log.OmitBodyPaths)
	assert.NotEmpty(t, cfg.FileSystem.Drive.CredentialsFile)
	assert.NotEmpty(t, cfg.Crypto.EncryptionKey)
	assert.NotEmpty(t, cfg.Apikeys)
	assert.NotEmpty(t, cfg.Services)

	// previous_keys is a slice of structs, which configKeys only checks as a single key;
	// assert the nested fields actually decode so a rotation example cannot silently rot.
	require.Len(t, cfg.JWT.PreviousKeys, 1)
	assert.NotEmpty(t, cfg.JWT.PreviousKeys[0].KeyID)
	assert.NotEmpty(t, cfg.JWT.PreviousKeys[0].AccessSecret)
	assert.NotEmpty(t, cfg.JWT.PreviousKeys[0].RefreshSecret)
}

// configKeys lists the dotted key of every struct field; maps count as a single key.
func configKeys(typ reflect.Type, prefix string) []string {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	var keys []string
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := field.Tag.Get("mapstructure")
		if name == "" {
			name = strings.ToLower(field.Name)
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct && fieldType.PkgPath() != "time" {
			keys = append(keys, configKeys(fieldType, prefix+name+".")...)
			continue
		}
		keys = append(keys, prefix+name)
	}
	return keys
}
