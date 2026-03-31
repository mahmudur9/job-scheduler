package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

type DB struct {
	Conn *sql.DB
}

func NewDb(conn string) (*DB, error) {
	db, err := sql.Open("sqlserver", conn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute) // After five minutes a single connection will be died and the connection will be replaced with a new one.

	maxRetries := 5
	baseDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			return &DB{Conn: db}, nil
		}

		time.Sleep(baseDelay * time.Duration(1<<i)) // exponential backoff
	}

	db.Close()
	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

func (d *DB) Close() error {
	return d.Conn.Close()
}
