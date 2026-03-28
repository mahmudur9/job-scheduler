package db

import (
	"database/sql"

	_ "github.com/denisenkom/go-mssqldb"
)

type DB struct {
	Conn *sql.DB
}

func NewDb(conn string) (*DB, error) {
	db, err := sql.Open("sqlserver", conn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{Conn: db}, nil
}
