package examples

import (
	"github.com/cyphix/databases/sql"
	"time"
)

func main() {
	// Initialize a new SQL database wrapper
	db := sql.NewDatabase().
		DatabaseType(sql.Postgres, true).
		DSN("host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai").
		MaxLifetime(time.Hour).
		SetMaxIdle(10).
		SetMaxOpen(100)

	// Open the connection
	_, err := db.Open()
	if err != nil {
		panic(err)
	}

	// Get the GORM client
	client, err := db.GetConnection()
	if err != nil {
		panic(err)
	}

	// Use the client...
	_ = client
}
