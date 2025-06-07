package config

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	gormMysql "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	SqlDB   *sql.DB
	PgxPool *pgxpool.Pool
	GDB     *gorm.DB
	Redis   *redis.Client
}

func NewDatabase(viper *viper.Viper, log *logrus.Logger) *DB {
	return &DB{
		// SqlDB:   mysqlDB(viper, log),
		PgxPool: postgresDB(viper, log),
		GDB:     gormDB("postgres", viper, log),
		Redis:   redisDB(viper, log),
	}
}

func postgresDB(viper *viper.Viper, log *logrus.Logger) *pgxpool.Pool {

	var pgx *pgxpool.Pool

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		viper.GetString("db.username"),
		viper.GetString("db.password"),
		viper.GetString("db.host"),
		viper.GetInt("db.port"),
		viper.GetString("db.name"),
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", "err")
	}
	config.MinConns = int32(viper.GetInt("db.min_open_connections"))
	config.MaxConns = int32(viper.GetInt("db.max_open_connections"))
	config.MaxConnLifetime = time.Second * time.Duration(viper.GetInt("db.connection_max_lifetime"))
	config.MaxConnIdleTime = time.Second * time.Duration(viper.GetInt("db.connection_max_idle_time"))
	config.HealthCheckPeriod = time.Second * time.Duration(viper.GetInt("db.health_check_period"))

	if pgx, err = pgxpool.NewWithConfig(context.Background(), config); err != nil {
		log.Fatalf("Unable to connect to postgres: %v", err)
	}

	if err = pgx.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping postgres: %v", err)
	}

	return pgx
}

func mysqlDB(viper *viper.Viper, log *logrus.Logger) *sql.DB {

	var db *sql.DB
	var err error

	cfg := mysql.Config{
		User:                 viper.GetString("db.username"),
		Passwd:               viper.GetString("db.password"),
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", viper.GetString("db.host"), viper.GetInt("db.port")),
		DBName:               viper.GetString("db.name"),
		AllowNativePasswords: true,
	}

	if db, err = sql.Open("mysql", cfg.FormatDSN()); err != nil {
		log.Fatalf("failed to connect to mysql: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping mysql: %v", err)
	}

	db.SetMaxOpenConns(viper.GetInt("db.max_open_connections"))
	db.SetConnMaxLifetime(time.Second * time.Duration(viper.GetInt("db.connection_max_lifetime")))
	db.SetConnMaxIdleTime(time.Second * time.Duration(viper.GetInt("db.connection_max_idle_time")))

	return db
}

func gormDB(driver string, viper *viper.Viper, log *logrus.Logger) *gorm.DB {
	var dialector gorm.Dialector
	usename := viper.GetString("db.username")
	password := viper.GetString("db.password")
	host := viper.GetString("db.host")
	port := viper.GetInt("db.port")
	dbName := viper.GetString("db.name")

	switch driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			host,
			usename,
			password,
			dbName,
			port,
		)
		dialector = gormPostgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&allowNativePasswords=true",
			usename,
			password,
			host,
			port,
			dbName,
		)
		dialector = gormMysql.Open(dsn)
	default:
		log.Fatalf("unsupported db driver: %s", driver)
	}

	gdb, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger: logger.New(&logrusWriter{Logger: log}, logger.Config{
			SlowThreshold:             time.Second * 5,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Warn,
		}),
	})
	if err != nil {
		log.Fatalf("failed to connect to gorm: %v", err)
	}

	connection, err := gdb.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm: %v", err)
	}

	connection.SetMaxOpenConns(viper.GetInt("db.max_open_connections"))
	connection.SetConnMaxLifetime(time.Second * time.Duration(viper.GetInt("db.connection_max_lifetime")))
	connection.SetConnMaxIdleTime(time.Second * time.Duration(viper.GetInt("db.connection_max_idle_time")))

	return gdb
}

func redisDB(viper *viper.Viper, log *logrus.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port")),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.db"),
		DialTimeout:  time.Second * time.Duration(viper.GetInt("redis.dial_timeout")),
		ReadTimeout:  time.Second * time.Duration(viper.GetInt("redis.read_timeout")),
		WriteTimeout: time.Second * time.Duration(viper.GetInt("redis.write_timeout")),
		PoolSize:     viper.GetInt("redis.pool_size"),
		PoolTimeout:  time.Second * time.Duration(viper.GetInt("redis.pool_timeout")),
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	return rdb
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	l.Logger.Tracef(message, args...)
}
