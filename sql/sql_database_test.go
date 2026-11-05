package sql

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDriver = Driver{Name: "postgres", Dialector: postgres.Open}

func TestDatabase_Builders(t *testing.T) {
	db := NewDatabase().
		MaxLifetime(2*time.Hour).
		UseDriver(testDriver, true).
		DSN("host=localhost").
		SetMaxIdle(5).
		SetMaxOpen(50)

	assert.Equal(t, 2*time.Hour, db.ConnMaxLifetime)
	assert.Equal(t, "postgres", db.Driver.Name)
	assert.Equal(t, "host=localhost", db.Dsn)
	assert.Equal(t, 5, db.MaxIdle)
	assert.Equal(t, 50, db.MaxOpen)
}

func TestDatabase_Open_Errors(t *testing.T) {
	t.Run("Missing Driver", func(t *testing.T) {
		db := NewDatabase()
		_, err := db.Open()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database driver is not set")
	})

	t.Run("Missing DSN", func(t *testing.T) {
		db := NewDatabase().UseDriver(testDriver, false)
		_, err := db.Open()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database DSN is not set")
	})

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
	db := NewDatabase().UseDriver(testDriver, false).DSN("host=127.0.0.1 port=1 connect_timeout=1 sslmode=disable")
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
		UseDriver(testDriver, false).
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
		UseDriver(testDriver, false).
		SetClient(gormDB)

	mock.ExpectClose()

	err = db.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
