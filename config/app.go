package config

import "os"

type AppVariable struct {
	Name         interface{}
	Env          interface{}
	Url          interface{}
	Timezone     interface{}
	DbConnection interface{}
	DbDriver     interface{}
	DbUrl        interface{}
	DbHost       interface{}
	DbPort       interface{}
	DbUser       interface{}
	DbPassword   interface{}
	DbName       interface{}
}

func SetAppVariable() {
	App = &AppVariable{
		Name:         env("APP_NAME", "Goilerplate"),
		Env:          env("APP_ENV", "production"),
		Url:          env("URL", "http://localhost:80"),
		Timezone:     env("TIMEZONE", "UTC"),
		DbConnection: env("DB_CONNECTION", "psql"),
		DbUrl:        env("DB_URL", ""),
		DbHost:       env("DB_HOST", "localhost"),
		DbPort:       env("DB_PORT", "5432"),
		DbUser:       env("DB_USER", "postgres"),
		DbPassword:   env("DB_PASSWORD", "postgres"),
		DbName:       env("DB_NAME", "goilerplate"),
	}
}

var App *AppVariable

func env(key string, defaultValue interface{}) interface{} {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
