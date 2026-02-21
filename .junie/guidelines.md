# go.databases - Guidelines

`go.databases` is a Go package for managing database connections and stores, supporting SQL (via GORM) and Redis. It is designed to be a lightweight wrapper that provides consistent logging and easy initialization for Go backends.

## Quick Reference

- **Build package**: `go build ./...`
- **Run tests**: `go test ./...`
- **Tidy modules**: `go mod tidy`

---

## Build & Dependencies

This package is built using **Go 1.26** and managed with Go Modules.

### Prerequisites
- **Go**: Version 1.26 or higher.

### Dependencies
The project uses several key libraries:
- **SQL ORM**: `gorm.io/gorm`
- **SQL Drivers**: `gorm.io/driver/postgres`, `gorm.io/driver/sqlite`
- **Redis Client**: `github.com/redis/go-redis/v9`
- **Logging**: `github.com/rs/zerolog` and `gitea.cyphix.dev/kade/go.logg` (internal wrapper)

---

## Testing Information

### Configuration
The project uses the standard library `testing` package along with:
- **`testify`**: for assertions (`assert` and `require`).
- **`DATA-DOG/go-sqlmock`**: for mocking SQL database interactions in `sql` package tests.
- **`alicebob/miniredis`**: for mocking Redis in `redis` package tests.

### Running Tests
- **All tests**: `go test ./...`
- **Verbose output**: `go test -v ./...`
- **Coverage**: `go test -cover ./...`

### Adding New Tests
Tests should be placed in the same package as the code they test, using the `_test.go` suffix. Use `sqlmock` for testing GORM-related logic to avoid dependency on a live database.

#### Example Test (with sqlmock)
```go
func TestDatabase_WithSqlmock(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	gormDB, _ := gorm.Open(dialector, &gorm.Config{})

	db := NewDatabase().
		DatabaseType(Postgres, false).
		SetClient(gormDB)

	mock.ExpectQuery("SELECT VERSION()").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 15.0"))

	conn, _ := db.GetConnection()
	assert.NotNil(t, conn)

	var version string
	err := conn.Raw("SELECT VERSION()").Scan(&version).Error
	assert.NoError(t, err)
	assert.Equal(t, "PostgreSQL 15.0", version)

	assert.NoError(t, mock.ExpectationsWereMet())
}
```

---

## Core Principles

Write code that is **performant, idiomatic, and maintainable**. Focus on clarity and consistency in logging and error handling.

### Modern Go Idioms (1.26+)
- **`any` over `interface{}`**: Always use `any` for generic interfaces.
- **Integer Loops**: Use `for i := range n` for simple count-based loops.
- **Slices & Maps Packages**: Leverage `slices` and `maps` (e.g., `slices.Contains`, `maps.Keys`).
- **JSON Struct Tags**: Use `omitzero` instead of `omitempty` for types like `time.Time`, `time.Duration`, structs, slices, and maps.
- **New Pattern**: Use `new(val)` for pointer assignments where applicable.

### Error Handling & Debugging
- **Error Wrapping**: Use `fmt.Errorf("...: %w", err)` to wrap errors.
- **Error Checking**: Use `errors.Is` and `errors.As` (or `errors.AsType[T](err)`) for checking specific errors.
- **Logging**: Use the `logger` field in the `Database` structs, which is initialized with `logg.Ctx`.

---

## Coding Standards & Best Practices

- **Builder Pattern**: The `sql.Database` uses a builder pattern for configuration (`DatabaseType`, `DSN`, `MaxLifetime`, etc.).
- **Consistency**: Maintain consistent logging patterns using `logg.KeyEvent` and `logg.KeyResult`.
- **Resource Management**: Always handle connection closing properly.

---

## Security & Privacy

To protect sensitive data and prevent accidental leakage of credentials to LLMs:
- **DO NOT READ `.env` files**: Never open or read the content of `.env` files or any other files containing secrets, API keys, or passwords.
- **Sensitive Data**: If you encounter files that appear to contain sensitive information (like private keys, certificates, or database credentials), avoid reading them unless absolutely necessary for the task.
