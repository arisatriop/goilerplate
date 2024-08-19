package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *pgxpool.Pool
var gdb *gorm.DB

type Con struct {
	Db  *pgxpool.Pool
	Gdb *gorm.DB
}

func CreateDBConnection() {
	db = SqlConnection()
	gdb = GormConnection()
	fmt.Println("Database: connected")
	fmt.Println()
}

func SqlConnection() *pgxpool.Pool {
	// exampleConnString := "postgres://username:password@host:post/dbname"

	app := GetAppVariable()

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		app.DbUser,
		app.DbPassword,
		app.DbHost,
		app.DbPort,
		app.DbName,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse config: %v\n", err)
		os.Exit(1)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), setPgxConfig(config))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to postgres: %v\n", err)
		os.Exit(1)
	}
	// defer conn.Close()

	fmt.Println("Postgres: ok")
	return conn
}

func setPgxConfig(config *pgxpool.Config) *pgxpool.Config {
	config.MaxConns = 50
	config.MinConns = 10
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 15 * time.Minute
	config.HealthCheckPeriod = time.Minute

	return config
}

func GormConnection() *gorm.DB {
	// exampleDSN := "host=localhost dbname=goilerplate password=postgres user=postgres port=5432 sslmode=disable TimeZone=Asia/Jakarta"

	// dsn := fmt.Sprintf(`
	// 	host=%s
	// 	dbname=%s
	// 	password=%s
	// 	user=%s
	// 	port=%s
	// 	sslmode=disable
	// 	TimeZone=Asia/Jakarta`,
	// 	App.DbHost,
	// 	App.DbName,
	// 	App.DbPassword,
	// 	App.DbUser,
	// 	App.DbPort,
	// )

	sqlDB := stdlib.OpenDB(*db.Config().ConnConfig)

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
		// DSN: dsn,
		// PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to gorm: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Gorm: ok")
	return gdb
}

func GetDBConnection() *Con {
	return &Con{
		Db:  db,
		Gdb: gdb,
	}
}
