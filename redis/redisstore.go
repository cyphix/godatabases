package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store represents the session store.
type Store struct {
	client *redis.Client
	prefix string
}

// NewRedisStore returns a new Store instance. The pool parameter should be a pointer
// to a go-redis connection pool.
func NewRedisStore(pool *redis.Client) *Store {
	return NewRedisStoreWithPrefix(pool, "scs:session:")
}

// NewRedisStoreWithPrefix returns a new Store instance. The pool parameter should be a pointer
// to a go-redis connection pool. The prefix parameter controls the Redis key
// prefix, which can be used to avoid naming clashes if necessary.
func NewRedisStoreWithPrefix(pool *redis.Client, prefix string) *Store {
	return &Store{
		client: pool,
		prefix: prefix,
	}
}

// Find returns the data for a given session token from the Store instance.
// If the session token is not found or is expired, the returned exists flag
// will be set to false.
func (r *Store) Find(token string) (b []byte, exists bool, err error) {
	// Set a timeout for the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Retrieve the value from Redis
	b, err = r.client.Get(ctx, r.prefix+token).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	return b, true, nil
}

// Commit adds a session token and data to the Store instance with the
// given expiry time. If the session token already exists then the data and
// expiry time are updated.
func (r *Store) Commit(token string, b []byte, expiry time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Using Pipelined to create a transaction-like operation
	_, err := r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, r.prefix+token, b, 0)         // Set the value
		pipe.PExpireAt(ctx, r.prefix+token, expiry) // Set expiration time
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Delete removes a session token and corresponding data from the Store
// instance.
func (r *Store) Delete(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.client.Del(ctx, r.prefix+token).Err()
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the Store instance.
func (r *Store) All() (map[string][]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	keys, err := r.client.Keys(ctx, r.prefix+"*").Result()
	if err != nil {
		return nil, err
	}

	sessions := make(map[string][]byte)

	for _, key := range keys {
		token := key[len(r.prefix):]

		data, exists, err := r.Find(token)
		if err != nil {
			return nil, err
		}

		if exists {
			sessions[token] = data
		}
	}

	return sessions, nil
}
