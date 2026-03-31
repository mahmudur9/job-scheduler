package migrations

import (
	"io/ioutil"
	"job-scheduler/internal/infrastructure/db"
	"log"
	"strings"
)

func create(dbConn string, path string) {
	database, _ := db.NewDb(dbConn)
	sqlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = database.Conn.Exec(string(sqlBytes))
	if err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}
	defer database.Close()
}

func RunMigration(dbConn string) {
	connArr := strings.Split(dbConn, "=")
	dbConnMaster := connArr[len(connArr)-2] + "=master"

	// Create database
	create(dbConnMaster, "internal/migrations/database_schema.sql")
	// Create tables
	create(dbConn, "internal/migrations/table_schema.sql")

	log.Println("Database migration applied successfully")
}
