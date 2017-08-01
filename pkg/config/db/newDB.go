package db

import (
	"database/sql"
	"fmt"
	"os"
)

// NewDB returns an open database handle of 0 or more underlying connections
func NewDB() (*sql.DB, error) {

	// Get Database connection credentials from environment variables
	dbName := os.Getenv("PG_DBNAME")
	dbUser := os.Getenv("PG_USERNAME")
	dbPassword := os.Getenv("PG_PASSWORD")

	// Craft string for database connection
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
