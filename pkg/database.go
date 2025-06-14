package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"goilerplate/config"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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

func NewDatabase(cfg *config.Config, log *logrus.Logger) *DB {
	return &DB{
		// SqlDB:   mysqlDB(viper, log),
		PgxPool: postgresDB(cfg, log),
		GDB:     gormDB("postgres", cfg, log),
		Redis:   redisDB(cfg, log),
	}
}

func postgresDB(cfg *config.Config, log *logrus.Logger) *pgxpool.Pool {

	var pgx *pgxpool.Pool

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", "err")
	}

	config.MinConns = int32(cfg.DB.MinOpenConnections)
	config.MaxConns = int32(cfg.DB.MaxOpenConnections)
	config.MaxConnLifetime = time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime)
	config.MaxConnIdleTime = time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime)
	config.HealthCheckPeriod = time.Second * time.Duration(cfg.DB.HealthCheckPeriod)

	if pgx, err = pgxpool.NewWithConfig(context.Background(), config); err != nil {
		log.Fatalf("Unable to connect to postgres: %v", err)
	}

	if err = pgx.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping postgres: %v", err)
	}

	return pgx
}

func mysqlDB(cfg *config.Config, log *logrus.Logger) *sql.DB {

	var db *sql.DB
	var err error

	config := mysql.Config{
		User:                 cfg.DB.Username,
		Passwd:               cfg.DB.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.DB.Host, cfg.DB.Port),
		DBName:               cfg.DB.Name,
		AllowNativePasswords: true,
	}

	if db, err = sql.Open("mysql", config.FormatDSN()); err != nil {
		log.Fatalf("failed to connect to mysql: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping mysql: %v", err)
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	db.SetConnMaxLifetime(time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime))
	db.SetConnMaxIdleTime(time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime))

	return db
}

func gormDB(driver string, cfg *config.Config, log *logrus.Logger) *gorm.DB {
	var dialector gorm.Dialector
	usename := cfg.DB.Username
	password := cfg.DB.Password
	host := cfg.DB.Host
	port := cfg.DB.Port
	dbName := cfg.DB.Name

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

	connection.SetMaxOpenConns(cfg.DB.MaxOpenConnections)
	connection.SetConnMaxLifetime(time.Second * time.Duration(cfg.DB.ConnectionMaxLifetime))
	connection.SetConnMaxIdleTime(time.Second * time.Duration(cfg.DB.ConnectionMaxIdleTime))

	return gdb
}

func redisDB(cfg *config.Config, log *logrus.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password, // no password set
		DB:           cfg.Redis.DB,       // use default DB
		DialTimeout:  time.Second * time.Duration(cfg.Redis.DialTimeout),
		ReadTimeout:  time.Second * time.Duration(cfg.Redis.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(cfg.Redis.WriteTimeout),
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  time.Second * time.Duration(cfg.Redis.PoolTimeout),
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
