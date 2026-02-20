package sql

import (
	"github.com/rs/zerolog/log"
)

type DbType int

const (
	Undefined DbType = iota
	SQLite
	Postgres
	MariaDb
)

func (dbType DbType) String() string {
	switch dbType {
	case SQLite:
		return "sqlite"
	case Postgres:
		return "postgres"
	case MariaDb:
		return "mariadb"
	case Undefined:
		return "undefined"
	default:
		log.Error().Msg("invalid database type")
		return "internal error"
	}
}
