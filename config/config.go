package config

import (
	"time"

	"goilerplate/pkg/filesystem"
)

type Config struct {
	App        App
	Server     Server
	DB         DB
	Redis      Redis
	JWT        JWT
	Log        *Logger
	FileSystem FileSystem
	Crypto     Crypto
	Apikey     []Apikey
}

type App struct {
	Env         string `mapstructure:"env"`
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Description string `mapstructure:"description"`
}

type Server struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Prefork      bool          `mapstructure:"prefork"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	EnableCORS   bool          `mapstructure:"enable_cors"`
	CORS         CORS
}

type CORS struct {
	AllowOrigin  string `mapstructure:"allow_origin"`
	AllowMethods string `mapstructure:"allow_methods"`
	AllowHeaders string `mapstructure:"allow_headers"`
}

type DB struct {
	Driver                string `mapstructure:"driver"`
	Host                  string `mapstructure:"host"`
	Port                  int    `mapstructure:"port"`
	Name                  string `mapstructure:"name"`
	SSLMode               string `mapstructure:"sslmode"`
	Username              string `mapstructure:"username"`
	Password              string `mapstructure:"password"`
	MinOpenConnections    int    `mapstructure:"min_open_connections"`
	MaxOpenConnections    int    `mapstructure:"max_open_connections"`
	ConnectionMaxLifetime int    `mapstructure:"connection_max_lifetime"`
	ConnectionMaxIdleTime int    `mapstructure:"connection_max_idle_time"`
	HealthCheckPeriod     int    `mapstructure:"health_check_period"`
}

type Redis struct {
	Enabled      bool          `mapstructure:"enabled"`
	Host         string        `mapstructure:"host"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
}

type JWT struct {
	SecretKey          string        `mapstructure:"secret_key"`
	AccessSecret       string        `mapstructure:"access_secret"`
	RefreshSecret      string        `mapstructure:"refresh_secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
	Issuer             string        `mapstructure:"issuer"`
}

type Logger struct {
	Level  string `mapstructure:"level"`
	Source bool   `mapstructure:"source"`
}

type FileSystem struct {
	Driver      string                 `mapstructure:"driver"`        // local, s3, drive
	MaxFileSize int64                  `mapstructure:"max_file_size"` // Maximum file size in bytes
	Local       filesystem.LocalConfig `mapstructure:"local"`
	S3          filesystem.S3Config    `mapstructure:"s3"`
	Drive       filesystem.DriveConfig `mapstructure:"drive"`
}

type Crypto struct {
	EncryptionKey string `mapstructure:"encryption_key"`
}

type Apikey struct {
	Name string `mapstructure:"name"`
	Key  string `mapstructure:"key"`
}
