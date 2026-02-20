package redis

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestStore_Operations(t *testing.T) {
	s := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	store := NewRedisStore(client)

	token := "test-token"
	data := []byte("session-data")
	expiry := time.Now().Add(time.Hour)

	// Test Commit
	err := store.Commit(token, data, expiry)
	assert.NoError(t, err)

	// Test Find (exists)
	foundData, exists, err := store.Find(token)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, data, foundData)

	// Test Find (not exists)
	_, exists, err = store.Find("non-existent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test All
	all, err := store.All()
	assert.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, data, all[token])

	// Test Delete
	err = store.Delete(token)
	assert.NoError(t, err)

	// Verify Delete
	_, exists, err = store.Find(token)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test All (empty)
	all, err = store.All()
	assert.NoError(t, err)
	assert.Len(t, all, 0)
}

func TestStore_Expiry(t *testing.T) {
	s := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	store := NewRedisStore(client)

	token := "expiry-token"
	data := []byte("expiry-data")
	expiry := time.Now().Add(time.Second)

	err := store.Commit(token, data, expiry)
	assert.NoError(t, err)

	// Fast-forward time in miniredis
	s.FastForward(2 * time.Second)

	foundData, exists, err := store.Find(token)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, foundData)
}

func TestNewRedisStoreWithPrefix(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	prefix := "custom:"
	store := NewRedisStoreWithPrefix(client, prefix)

	token := "token"
	data := []byte("data")
	err := store.Commit(token, data, time.Now().Add(time.Hour))
	assert.NoError(t, err)

	// Check if miniredis has the key with prefix
	assert.True(t, s.Exists(prefix+token))
}

func TestStore_ErrorConditions(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	store := NewRedisStore(client)

	// Close the client to force errors
	client.Close()

	// Test Find error
	_, _, err := store.Find("token")
	assert.Error(t, err)

	// Test Commit error
	err = store.Commit("token", []byte("data"), time.Now().Add(time.Hour))
	assert.Error(t, err)

	// Test All error
	_, err = store.All()
	assert.Error(t, err)

	// Test Delete error
	err = store.Delete("token")
	assert.Error(t, err)
}
