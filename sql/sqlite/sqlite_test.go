package sqlite_test

import (
	"bytes"
	"path/filepath"
	"testing"

	logg "github.com/cyphix/gologg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cyphix/godatabases/sql"
	"github.com/cyphix/godatabases/sql/sqlite"
)

func TestOpen(t *testing.T) {
	db, err := sql.NewDatabase().UseDriver(sqlite.Driver, false).DSN(filepath.Join(t.TempDir(), "test.db")).Open()
	require.NoError(t, err)

	conn, err := db.GetConnection()
	require.NoError(t, err)
	assert.Equal(t, "sqlite", conn.Name())
	assert.NoError(t, db.Close())
}

func TestOpen_GormError(t *testing.T) {
	_, err := sql.NewDatabase().UseDriver(sqlite.Driver, false).DSN("/nonexistent/path/to/db").Open()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "opening gorm connection")
}

func TestLoggerSuppression(t *testing.T) {
	var buf bytes.Buffer
	logg.InitWithWriters("debug", &buf)
	defer logg.Init("info")

	dir := t.TempDir()

	quiet, err := sql.NewDatabase().UseDriver(sqlite.Driver, false).DSN(filepath.Join(dir, "quiet.db")).Open()
	require.NoError(t, err)
	require.NoError(t, quiet.Close())
	assert.Empty(t, buf.String())

	loud, err := sql.NewDatabase().UseDriver(sqlite.Driver, true).DSN(filepath.Join(dir, "loud.db")).Open()
	require.NoError(t, err)
	require.NoError(t, loud.Close())
	assert.NotEmpty(t, buf.String())
}
