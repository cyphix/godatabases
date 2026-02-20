package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/cyphix/logg"
)

type Database struct {
	client *redis.Client
	logger zerolog.Logger
}

func NewRedisDatabase(
	host string, port string, password string, databaseNumber int, useLogger bool,
) (*Database, error) {
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}

	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     host + ":" + port,
			Password: password,
			DB:       databaseNumber,
		},
	)

	redisDb := &Database{
		client: redisClient,
	}

	// Create a sub-zlog
	redisDb.logger = logg.Ctx("databases", "redis")
	if useLogger {
		redisDb.logger.Level(zerolog.GlobalLevel())
	} else {
		redisDb.logger.Level(zerolog.Disabled)
	}

	// Check the connection to make sure that it works
	err := redisDb.ConnectionCheck()

	return redisDb, err
}

func (db *Database) Close() error {
	if db.client != nil {
		err := db.client.Close()
		if err != nil {
			db.logger.Error().Stack().Err(err).
				Str(logg.KeyEvent, "close").
				Str(logg.KeyResult, "fail").
				Msg("Redis connection failed to closed")
			return err
		}

		db.logger.Info().
			Str(logg.KeyEvent, "close").
			Str(logg.KeyResult, "success").
			Msg("Redis connection closed")
	}

	return nil
}

func (db *Database) ConnectionCheck() error {
	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test the connection
	result, err := db.client.Ping(ctx).Result()
	if err != nil {
		db.logger.Warn().Err(err).
			Str(logg.KeyEvent, "connect").
			Str(logg.KeyResult, "fail").
			Msgf("Failed to connect to the database: %s", result)
		return err
	}

	db.logger.Info().
		Str(logg.KeyEvent, "connect").
		Str(logg.KeyResult, "success").
		Msgf("Successfully connected to Redis: %s", result)
	return nil
}

func (db *Database) CreateRedisStore() (*Store, error) {
	if db.client == nil {
		return nil, errors.New("redis client is not initialized")
	}
	store := NewRedisStore(db.client)
	return store, nil
}

func (db *Database) Client() *redis.Client {
	return db.client
}
