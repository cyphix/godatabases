# godatabases

`go.databases` is a Go package for managing database connections and stores, supporting SQL (via GORM) and Redis. It is designed to be a lightweight wrapper that provides consistent logging and easy initialization for Go backends.

## Features

- **SQL Database Management**: Supports SQLite and PostgreSQL via GORM with a builder pattern.
- **Redis Integration**: Easy initialization for Redis clients.
- **Redis Store**: A session store implementation for Redis.
- **Consistent Logging**: Integrated with `zerolog` and `logg` for structured logging.
- **Modern Go**: Built with Go 1.26 idioms.

## Installation

```bash
go get github.com/cyphix/godatabases
```

## Quick Start

### SQL Database (PostgreSQL)

```go
import (
    "github.com/cyphix/godatabases/sql"
)

db, err := sql.NewDatabase().
    DatabaseType(sql.Postgres, true).
    DSN("host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable").
    Open()

if err != nil {
    panic(err)
}
defer db.Close()

conn, err := db.GetConnection()
// Use conn (*gorm.DB) for your queries
```

### Redis Database

```go
import (
    "github.com/cyphix/godatabases/redis"
)

db, err := redis.NewRedisDatabase("localhost", "6379", "", 0, true)
if err != nil {
    panic(err)
}
defer db.Close()

client := db.Client()
// Use client (*redis.Client) for your operations
```

### Redis Store

```go
store, err := db.CreateRedisStore()
if err != nil {
    panic(err)
}

// Use store for session management
err = store.Commit("my-token", []byte("session-data"), time.Now().Add(time.Hour))
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
