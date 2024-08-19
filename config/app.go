package config

import (
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
)

type AppVariable struct {
	Name          interface{}
	Env           interface{}
	Url           interface{}
	Timezone      interface{}
	DbConnection  interface{}
	DbDriver      interface{}
	DbUrl         interface{}
	DbHost        interface{}
	DbPort        interface{}
	DbUser        interface{}
	DbPassword    interface{}
	DbName        interface{}
	ElasticHost   interface{}
	ElasticPort   interface{}
	RedisHost     interface{}
	RedisPort     interface{}
	RedisPassword interface{}
	DB            *Con
	RedisClient   *redis.Client
	ElasticClient *elasticsearch.Client
}

func SetAppVariable() *AppVariable {
	app = &AppVariable{
		Name:          env("APP_NAME", "Goilerplate"),
		Env:           env("APP_ENV", "production"),
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

func env(key string, defaultValue interface{}) interface{} {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
