# Architecture & Layout

`go.databases` follows a clean, package-based architecture where different database types are separated into their own packages.

## Directory Structure

- **`sql/`**: Contains SQL database management logic.
  - **`dbtype.go`**: Defines supported SQL database types (SQLite, Postgres, MariaDb).
  - **`sql_database.go`**: The main `Database` struct and its builder methods for GORM connections.
- **`redis/`**: Contains Redis-related logic.
  - **`redis_database.go`**: Connection management for Redis.
  - **`redisstore.go`**: Implementation of a Redis-backed session store (compatible with `scs`).

## Patterns

### SQL Builder Pattern

The `sql.Database` struct uses a builder pattern for easy and readable configuration:

```go
db := sql.NewDatabase().
    DatabaseType(sql.Postgres, true).
    DSN("host=localhost ...").
    SetMaxOpen(100)

gormConn, err := db.Open()
```

### Logging

Both `sql` and `redis` packages use a unified logging approach via the `logg` package, ensuring that database events are logged consistently across the application.

### Examples

The `.junie/examples/` directory contains usage examples for both SQL and Redis database wrappers.
