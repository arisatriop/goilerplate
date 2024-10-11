package config

import (
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

var app *App

type App struct {
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
	Validator     *validator.Validate
	DB            *Con
	RedisClient   *redis.Client
	ElasticClient *elasticsearch.Client
}

func GetAppVariable() *App {
	return app
}

func env(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SetAppVariable() *App {
	app = &App{
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
		CacheDriver:   env("CACHE_DRIVER", "file"), // available: redis
		LogChannel:    env("LOG_CHANNEL", "file"),  // available: file, elastic
		RedisHost:     env("REDIS_HOST", "localhost"),
		RedisPort:     env("REDIS_PORT", "6379"),
		RedisPassword: env("REDIS_PASSWORD", "secret"),
	}

	return app
}

func (app *App) SetValidator() {
	app.Validator = validator.New()
}

func (app *App) SetDBConnection() {
	app.DB = CreateDBConnection()
}

func (app *App) SetRedisConnection() {
	if app.CacheDriver == "redis" {
		app.RedisClient = CreateRedisConnection()
	}
}

func (app *App) SetElasticConnection() {
	if app.LogChannel == "elastic" {
		app.ElasticClient = CreateElasticConnection()
	}
}
