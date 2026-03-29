package migrations

import (
	"io/ioutil"
	"job-scheduler/internal/infrastructure/db"
	"log"
)

func RunMigration(db *db.DB) {
	sqlBytes, err := ioutil.ReadFile("internal/migrations/schema.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = db.Conn.Exec(string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	log.Println("Database migration applied successfully")
}
