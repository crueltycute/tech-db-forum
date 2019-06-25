package service

import "github.com/jackc/pgx"

type txOrDb interface {
	QueryRow(string, ...interface{}) *pgx.Row
	Query(string, ...interface{}) (*pgx.Rows, error)
	Exec(string, ...interface{}) (commandTag pgx.CommandTag, err error)
}

type dbOrConn interface {
	QueryRow(string, ...interface{}) *pgx.Row
	Query(string, ...interface{}) (*pgx.Rows, error)
	Exec(string, ...interface{}) (commandTag pgx.CommandTag, err error)
	Begin() (*pgx.Tx, error)
}
