package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyphix/godatabases/sql/postgres"
)

func TestDriver(t *testing.T) {
	assert.Equal(t, "postgres", postgres.Driver.Name)
	assert.Equal(t, "postgres", postgres.Driver.Dialector("host=localhost").Name())
}
