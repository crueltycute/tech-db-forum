package db

import (
	"github.com/jackc/pgx"
)

var Connection *pgx.ConnPool

func EnsureDBConnection(conf pgx.ConnConfig) error {
	var err error

	Connection, err = pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     conf,
			MaxConnections: 50,
		})

	return err
}
