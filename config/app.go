package config

import (
	"fmt"
	"os"
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
}

func SetAppVariable() {
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

	fmt.Println("App Name: ", app.Name)
	fmt.Println("App Env: ", app.Env)
	fmt.Println("App Url: ", app.Url)
	fmt.Println("App Timezone: ", app.Timezone)
	fmt.Println("Db Connection: ", app.DbConnection)
	fmt.Println("Db Url: ", app.DbUrl)
	fmt.Println("Db Host: ", app.DbHost)
	fmt.Println("Db Port: ", app.DbPort)
	fmt.Println("Db User: ", app.DbUser)
	fmt.Println("Db Password: ", app.DbPassword)
	fmt.Println("Db Name: ", app.DbName)
	fmt.Println("Elastic Host: ", app.ElasticHost)
	fmt.Println("Elastic Port: ", app.ElasticPort)
	fmt.Println("Redis Host: ", app.RedisHost)
	fmt.Println("Redis Port: ", app.RedisPort)
	fmt.Println("Redis Password: ", app.RedisPassword)
	fmt.Println()
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
