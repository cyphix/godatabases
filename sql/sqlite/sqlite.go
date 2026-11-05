// Package sqlite provides the SQLite driver for godatabases/sql.
package sqlite

import (
	gormsqlite "gorm.io/driver/sqlite"

	"github.com/cyphix/godatabases/sql"
)

// Driver opens SQLite databases. The DSN is a file path such as "app.db".
var Driver = sql.Driver{Name: "sqlite", Dialector: gormsqlite.Open}
