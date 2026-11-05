// Package postgres provides the PostgreSQL driver for godatabases/sql.
package postgres

import (
	gormpostgres "gorm.io/driver/postgres"

	"github.com/cyphix/godatabases/sql"
)

// Driver opens PostgreSQL connections. DSN example:
//
//	"host=localhost user=username password=passwd dbname=dbname port=5432 sslmode=disable TimeZone=America/Los_Angeles"
var Driver = sql.Driver{Name: "postgres", Dialector: gormpostgres.Open}
