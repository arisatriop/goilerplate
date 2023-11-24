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
	// exampleConnString := "postgres://username:password@host:post/dbname"

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		App.DbUser,
		App.DbPassword,
		App.DbHost,
		App.DbPort,
		App.DbName,
	)

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
	// exampleDSN := "host=localhost dbname=goilerplate password=postgres user=postgres port=5432 sslmode=disable TimeZone=Asia/Jakarta"

	dsn := fmt.Sprintf(`
		host=%s
		dbname=%s
		password=%s
		user=%s
		port=%s
		sslmode=disable
		TimeZone=Asia/Jakarta`,
		App.DbHost,
		App.DbName,
		App.DbPassword,
		App.DbUser,
		App.DbPort,
	)

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
