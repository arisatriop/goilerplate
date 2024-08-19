package config

import (
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
)

type AppVariable struct {
	Name          string
	Env           string
	Url           string
	Timezone      string
	DbConnection  string
	DbDriver      string
	DbUrl         string
	DbHost        string
	DbPort        string
	DbUser        string
	DbPassword    string
	DbName        string
	ElasticHost   string
	ElasticPort   string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	CacheDriver   string
	LogChannel    string
	DB            *Con
	RedisClient   *redis.Client
	ElasticClient *elasticsearch.Client
}

func SetAppVariable() *AppVariable {
	app = &AppVariable{
		Name:          env("APP_NAME", "Goilerplate"),
		Env:           env("APP_ENV", "local"),
		Url:           env("URL", "http://localhost:80"),
		Timezone:      env("TIMEZONE", "UTC"),
		DbConnection:  env("DB_CONNECTION", "psql"),
		DbUrl:         env("DB_URL", ""),
		DbHost:        env("DB_HOST", "localhost"),
		DbPort:        env("DB_PORT", "5432"),
		DbUser:        env("DB_USER", "postgres"),
		DbPassword:    env("DB_PASSWORD", "postgres"),
		DbName:        env("DB_NAME", "goilerplate"),
		ElasticHost:   env("ELASTIC_HOST", "http://localhost"),
		ElasticPort:   env("ELASTIC_PORT", "9200"),
		CacheDriver:   env("CACHE_DRIVER", "file"),
		LogChannel:    env("LOG_CHANNEL", "file"),
		RedisHost:     env("REDIS_HOST", "localhost"),
		RedisPort:     env("REDIS_PORT", "6379"),
		RedisPassword: env("REDIS_PASSWORD", "secret"),
	}

	return app
}

func GetAppVariable() *AppVariable {
	return app
}

var app *AppVariable

func env(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
