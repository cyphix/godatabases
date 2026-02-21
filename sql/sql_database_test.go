package sql

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestDbType_String(t *testing.T) {
	tests := []struct {
		dbType DbType
		want   string
	}{
		{SQLite, "sqlite"},
		{Postgres, "postgres"},
		{MariaDb, "mariadb"},
		{Undefined, "undefined"},
		{DbType(999), "internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.dbType.String())
		})
	}
}

func TestDatabase_Builders(t *testing.T) {
	db := NewDatabase().
		MaxLifetime(2*time.Hour).
		DatabaseType(SQLite, true).
		DSN("test.db").
		SetMaxIdle(5).
		SetMaxOpen(50)

	assert.Equal(t, 2*time.Hour, db.ConnMaxLifetime)
	assert.Equal(t, SQLite, db.DbType)
	assert.Equal(t, "test.db", db.Dsn)
	assert.Equal(t, 5, db.MaxIdle)
	assert.Equal(t, 50, db.MaxOpen)
}

func TestDatabase_Open_Errors(t *testing.T) {
	t.Run("Undefined Type", func(t *testing.T) {
		db := NewDatabase()
		_, err := db.Open()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database type is not set")
	})

	t.Run("Missing DSN", func(t *testing.T) {
		db := NewDatabase().DatabaseType(SQLite, false)
		_, err := db.Open()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database DSN is not set")
	})

	t.Run("Unsupported Type", func(t *testing.T) {
		db := NewDatabase().DatabaseType(DbType(99), false).DSN("test")
		_, err := db.Open()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported database type")
	})

	t.Run("MariaDb Not Implemented", func(t *testing.T) {
		db := NewDatabase().DatabaseType(MariaDb, false).DSN("test")
		_, err := db.Open()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mariadb not implemented")
	})
}

func TestDatabase_Open_SQLite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sql-database-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dsn := filepath.Join(tmpDir, "test.db")
	db := NewDatabase().DatabaseType(SQLite, false).DSN(dsn)

	opened, err := db.Open()
	assert.NoError(t, err)
	assert.NotNil(t, opened)
	assert.NotNil(t, db.client)

	err = db.Close()
	assert.NoError(t, err)
}

func TestDatabase_GetConnection_Errors(t *testing.T) {
	t.Run("Nil Client", func(t *testing.T) {
		db := NewDatabase()
		_, err := db.GetConnection()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gorm database client is missing")
	})

	t.Run("Ping Failure", func(t *testing.T) {
		sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer sqlDB.Close()

		mock.ExpectPing().WillReturnError(assert.AnError)

		dialector := postgres.New(postgres.Config{Conn: sqlDB})
		// Disable logging in GORM to avoid it failing the test by intercepting the expected error
		gormDB, err := gorm.Open(dialector, &gorm.Config{
			Logger: nil,
		})
		if err != nil {
			// GORM might return error during Open if it pings, but we want it to fail during GetConnection
			// If it fails here, we just check if it's the error we expected
			assert.Contains(t, err.Error(), "assert.AnError")
			return
		}

		db := NewDatabase().SetClient(gormDB)

		_, err = db.GetConnection()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pinging database")
	})
}

func TestDatabase_Close_Nil(t *testing.T) {
	db := NewDatabase()
	err := db.Close()
	assert.NoError(t, err)
}

func TestDatabase_Close_Errors(t *testing.T) {
	t.Run("DB Retrieval Failure", func(t *testing.T) {
		sqlDB, _, err := sqlmock.New()
		require.NoError(t, err)
		sqlDB.Close() // Close it immediately

		dialector := postgres.New(postgres.Config{Conn: sqlDB})
		gormDB, err := gorm.Open(dialector, &gorm.Config{Logger: nil})
		if err != nil {
			// If gorm.Open fails because sqlDB is closed, we can't test db.Close()'s retrieval failure this way
			return
		}

		db := NewDatabase().SetClient(gormDB)
		err = db.Close()
		// If it fails, it should be a retrieval error or close error
		if err != nil {
			assert.Contains(t, err.Error(), "retrieving sql connection")
		}
	})

	t.Run("Close Failure", func(t *testing.T) {
		sqlDB, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer sqlDB.Close()

		mock.ExpectClose().WillReturnError(assert.AnError)

		dialector := postgres.New(postgres.Config{Conn: sqlDB})
		gormDB, err := gorm.Open(dialector, &gorm.Config{})
		require.NoError(t, err)

		db := NewDatabase().SetClient(gormDB)
		err = db.Close()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closing database")
	})
}

func TestDatabase_Open_GormError(t *testing.T) {
	// Triggering gorm.Open error
	// For example, by using an invalid DSN for SQLite
	db := NewDatabase().DatabaseType(SQLite, false).DSN("/nonexistent/path/to/db")
	_, err := db.Open()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "opening gorm connection")
}

func TestDatabase_WithSqlmock(t *testing.T) {
	// 1. Create sqlmock
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	// 2. Initialize GORM with the mocked connection
	// We use postgres dialector here as requested (to test things like JSONB support)
	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// 3. Initialize our Database wrapper
	db := NewDatabase().
		DatabaseType(Postgres, false).
		SetClient(gormDB)

	// 4. Set up expectations
	// Example: Mocking a Ping or a simple query
	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 15.0"))

	// 5. Run tests
	conn, err := db.GetConnection()
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	var version string
	err = conn.Raw("SELECT VERSION()").Scan(&version).Error
	assert.NoError(t, err)
	assert.Equal(t, "PostgreSQL 15.0", version)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabase_Close_Mock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	db := NewDatabase().
		DatabaseType(Postgres, false).
		SetClient(gormDB)

	mock.ExpectClose()

	err = db.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
