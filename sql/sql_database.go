package sql

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	logg "github.com/cyphix/gologg"
)

type Database struct {
	client          *gorm.DB
	DbType          DbType        `json:"db_type" omitzero:"true"`
	Dsn             string        `json:"dsn" omitzero:"true"`
	MaxIdle         int           `json:"max_idle" omitzero:"true"`
	MaxOpen         int           `json:"max_open" omitzero:"true"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" omitzero:"true"`

	logger zerolog.Logger
}

func NewDatabase() *Database {
	db := &Database{
		DbType:          Undefined,
		MaxIdle:         10,
		MaxOpen:         100,
		ConnMaxLifetime: time.Hour,
	}

	db.logger = logg.Ctx("databases", "sql").With().Str("db_type", db.DbType.String()).Logger()
	db.logger.Level(zerolog.Disabled)

	return db
}

// Builder methods

func (db *Database) MaxLifetime(duration time.Duration) *Database {
	db.ConnMaxLifetime = duration
	return db
}

func (db *Database) DatabaseType(dbType DbType, useLogger bool) *Database {
	db.DbType = dbType

	db.logger = logg.Ctx("databases", "sql").With().Str("db_type", dbType.String()).Logger()
	if useLogger {
		db.logger.Level(zerolog.GlobalLevel())
	} else {
		db.logger.Level(zerolog.Disabled)
	}

	return db
}

// DSN
//
// Different dsn supported database types:
//
//	sqlite: "filename.db"
//	postgres: "host=localhost user=username password=passwd dbname=dbname port=9920 sslmode=disable TimeZone=America/Los_Angeles"
//	mariadb: "username:passwd@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
func (db *Database) DSN(dsn string) *Database {
	db.Dsn = dsn
	return db
}

func (db *Database) SetMaxIdle(number int) *Database {
	db.MaxIdle = number
	return db
}

func (db *Database) SetMaxOpen(number int) *Database {
	db.MaxOpen = number
	return db
}

func (db *Database) SetClient(client *gorm.DB) *Database {
	db.client = client
	return db
}

// Execution methods

func (db *Database) Close() error {
	if db.client == nil {
		return nil
	}

	sqlDB, err := db.client.DB()
	if err != nil {
		db.logger.Error().Err(err).
			Str(logg.KeyEvent, "retrieval").
			Str(logg.KeyResult, "fail").
			Msg("failed to retrieve sql connection from gorm")
		return fmt.Errorf("retrieving sql connection: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		db.logger.Error().Err(err).
			Str(logg.KeyEvent, "close").
			Str(logg.KeyResult, "fail").
			Msg("failed to close database connection")
		return fmt.Errorf("closing database: %w", err)
	}

	db.logger.Info().
		Str(logg.KeyEvent, "close").
		Str(logg.KeyResult, "success").
		Msg("database connection closed")

	return nil
}

func (db *Database) GetConnection() (*gorm.DB, error) {
	if db.client == nil {
		db.logger.Error().
			Str(logg.KeyEvent, "connect").
			Str(logg.KeyResult, "fail").
			Msg("gorm client is nil")
		return nil, errors.New("gorm database client is missing")
	}

	sqlDB, err := db.client.DB()
	if err != nil {
		db.logger.Error().Err(err).
			Str(logg.KeyEvent, "retrieval").
			Str(logg.KeyResult, "fail").
			Msg("failed to retrieve sql connection from gorm")
		return nil, fmt.Errorf("retrieving sql connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		db.logger.Warn().Err(err).
			Str(logg.KeyEvent, "ping").
			Str(logg.KeyResult, "fail").
			Msg("database ping failed")
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db.client, nil
}

func (db *Database) Open() (*Database, error) {
	if db.DbType == Undefined {
		db.logger.Error().
			Str(logg.KeyEvent, "open").
			Str(logg.KeyResult, "fail").
			Msg("database type not set")
		return nil, errors.New("database type is not set")
	}
	if db.Dsn == "" {
		db.logger.Error().
			Str(logg.KeyEvent, "open").
			Str(logg.KeyResult, "fail").
			Msg("database DSN not set")
		return nil, errors.New("database DSN is not set")
	}

	var dialector gorm.Dialector
	switch db.DbType {
	case SQLite:
		dialector = sqlite.Open(db.Dsn)
	case Postgres:
		dialector = postgres.Open(db.Dsn)
	case MariaDb:
		return nil, errors.New("mariadb not implemented")
	default:
		db.logger.Error().
			Str(logg.KeyEvent, "open").
			Str(logg.KeyResult, "fail").
			Msgf("unsupported database type: %s", db.DbType)
		return nil, fmt.Errorf("unsupported database type: %s", db.DbType)
	}

	sqlDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		db.logger.Error().Err(err).
			Str(logg.KeyEvent, "open").
			Str(logg.KeyResult, "fail").
			Msg("failed to open gorm connection")
		return nil, fmt.Errorf("opening gorm connection: %w", err)
	}

	db.client = sqlDB

	conn, err := db.client.DB()
	if err != nil {
		db.logger.Error().Err(err).
			Str(logg.KeyEvent, "retrieval").
			Str(logg.KeyResult, "fail").
			Msg("failed to retrieve sql connection after open")
		return nil, fmt.Errorf("retrieving sql connection after open: %w", err)
	}

	conn.SetMaxIdleConns(db.MaxIdle)
	conn.SetMaxOpenConns(db.MaxOpen)
	conn.SetConnMaxLifetime(db.ConnMaxLifetime)

	db.logger.Info().
		Str(logg.KeyEvent, "open").
		Str(logg.KeyResult, "success").
		Msg("database opened successfully")

	return db, nil
}
