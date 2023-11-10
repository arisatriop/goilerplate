package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *pgx.Conn
var gdb *gorm.DB

type Con struct {
	Db  *pgx.Conn
	Gdb *gorm.DB
	Dtx *pgx.Tx
	Gtx *gorm.Tx
}

func CreateDBConnection() {
	db = SqlConnection()
	gdb = GormConnection()
	fmt.Println("Connected to database")
}

func SqlConnection() *pgx.Conn {
	// Example connection url
	connString := "postgres://username:password@host:post/dbname"

	connString = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	connString = fmt.Sprintf("postgres://arisatrio:%s@localhost:5432/goilerplate", "")
	db, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close(context.Background())

	fmt.Println("Postgres ok")
	return db
}

func GormConnection() *gorm.DB {
	// Example connection url
	dsn := "host=localhost dbname=goilerplate password=postgres user=postgres port=5432 sslmode=disable TimeZone=Asia/Jakarta"

	dsn = fmt.Sprintf(`
		host=%s
		dbname=%s
		password=%s
		user=%s
		port=%s
		sslmode=disable
		TimeZone=Asia/Jakarta`,
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PORT"),
	)

	dsn = fmt.Sprintf("host=localhost dbname=goilerplate password=%s user=arisatrio port=5432 sslmode=disable TimeZone=Asia/Jakarta", "")
	gdb, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
		// PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to gorm: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Gorm ok")
	return gdb
}

func GetDBConnection() *Con {
	return &Con{
		Db:  db,
		Gdb: gdb,
		Dtx: nil,
		Gtx: nil,
	}
}
