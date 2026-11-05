package sql

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	logg "github.com/cyphix/gologg"
)

type Database struct {
	client          *gorm.DB
	Driver          Driver        `json:"-"`
	Dsn             string        `json:"dsn" omitzero:"true"`
	MaxIdle         int           `json:"max_idle" omitzero:"true"`
	MaxOpen         int           `json:"max_open" omitzero:"true"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" omitzero:"true"`

	logger zerolog.Logger
}

func NewDatabase() *Database {
	db := &Database{
		MaxIdle:         10,
		MaxOpen:         100,
		ConnMaxLifetime: time.Hour,
	}

	db.logger = logg.Ctx("databases", "sql").Level(zerolog.Disabled)

	return db
}

// Builder methods

func (db *Database) MaxLifetime(duration time.Duration) *Database {
	db.ConnMaxLifetime = duration
	return db
}

// UseDriver selects the backend, e.g. postgres.Driver from
// github.com/cyphix/godatabases/sql/postgres or sqlite.Driver from
// github.com/cyphix/godatabases/sql/sqlite.
func (db *Database) UseDriver(driver Driver, useLogger bool) *Database {
	db.Driver = driver

	db.logger = logg.Ctx("databases", "sql").With().Str("driver", driver.Name).Logger()
	if useLogger {
		db.logger = db.logger.Level(zerolog.GlobalLevel())
	} else {
		db.logger = db.logger.Level(zerolog.Disabled)
	}

	return db
}

// DSN sets the connection string; its format is defined by the driver.
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
	if db.Driver.Dialector == nil {
		db.logger.Error().
			Str(logg.KeyEvent, "open").
			Str(logg.KeyResult, "fail").
			Msg("database driver not set")
		return nil, errors.New("database driver is not set")
	}
	if db.Dsn == "" {
		db.logger.Error().
			Str(logg.KeyEvent, "open").
			Str(logg.KeyResult, "fail").
			Msg("database DSN not set")
		return nil, errors.New("database DSN is not set")
	}

	sqlDB, err := gorm.Open(db.Driver.Dialector(db.Dsn), &gorm.Config{})
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
