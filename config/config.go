package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App    App
	Server Server
	DB     DB
	Redis  Redis
	JWT    JWT
	Log    Log
}

type App struct {
	Name        string
	Version     string
	Description string
	AppURL      string `mapstructure:"app_url"`
}

type Server struct {
	Host         string
	Port         int
	Prefork      bool
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	EnableCORS   bool
	CORS         CORS
}

type CORS struct {
	AllowOrigin  string `mapstructure:"allow_origin"`
	AllowMethods string `mapstructure:"allow_methods"`
	AllowHeaders string `mapstructure:"allow_headers"`
}

type DB struct {
	Host                  string
	Port                  int
	Name                  string
	SSLMode               string `mapstructure:"sslmode"`
	Username              string
	Password              string
	MinOpenConnections    int `mapstructure:"min_open_connections"`
	MaxOpenConnections    int `mapstructure:"max_open_connections"`
	ConnectionMaxLifetime int `mapstructure:"connection_max_lifetime"`
	ConnectionMaxIdleTime int `mapstructure:"connection_max_idle_time"`
	HealthCheckPeriod     int `mapstructure:"health_check_period"`
}

type Redis struct {
	Host         string
	Port         int
	Password     string
	DB           int
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
}

type JWT struct {
	Secret             string
	RefreshSecret      string `mapstructure:"refresh_secret"`
	AccessTokenExpiry  int    `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry int    `mapstructure:"refresh_token_expiry"`
}

type Log struct {
	Level  int
	Output string
}

func Load(v *viper.Viper) *Config {

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct, %w", err))
	}

	return &cfg
}
