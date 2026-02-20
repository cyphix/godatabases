package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisDatabase(t *testing.T) {
	s := miniredis.RunT(t)

	// Test with miniredis
	db, err := NewRedisDatabase(s.Host(), s.Port(), "", 0, false)
	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, db.Client())

	err = db.ConnectionCheck()
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	// Test with logger enabled
	dbLogger, err := NewRedisDatabase(s.Host(), s.Port(), "", 0, true)
	assert.NoError(t, err)
	assert.NotNil(t, dbLogger)
	dbLogger.Close()

	// Test with default host/port (will fail if no redis on localhost:6379, but we test the branch)
	// Actually, we can't easily test the default host/port branch without a running redis on 6379.
	// But we can check that it doesn't panic if we pass empty strings.
	_, _ = NewRedisDatabase("", "", "", 0, false)
}

func TestNewRedisDatabase_Fail(t *testing.T) {
	// Point to a non-existent port
	_, err := NewRedisDatabase("localhost", "1", "", 0, false)
	assert.Error(t, err)
}

func TestDatabase_CreateRedisStore(t *testing.T) {
	s := miniredis.RunT(t)

	db, err := NewRedisDatabase(s.Host(), s.Port(), "", 0, false)
	require.NoError(t, err)

	store, err := db.CreateRedisStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Test nil client check
	dbNil := &Database{client: nil}
	_, err = dbNil.CreateRedisStore()
	assert.Error(t, err)
	assert.Equal(t, "redis client is not initialized", err.Error())
}

func TestDatabase_Close_NilClient(t *testing.T) {
	db := &Database{client: nil}
	err := db.Close()
	assert.NoError(t, err)
}
