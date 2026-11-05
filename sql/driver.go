package sql

import "gorm.io/gorm"

// Driver selects the GORM dialector used by Database.Open. Driver values are
// provided by the backend subpackages (sql/postgres, sql/sqlite) so callers
// only link the database drivers they import.
type Driver struct {
	Name      string
	Dialector func(dsn string) gorm.Dialector
}
