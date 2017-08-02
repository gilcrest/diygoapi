// Package env has a type to store common environment related items
// sql db, logger, etc. as well as a constructor-like function to instantiate it
package env

import (
	"database/sql"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Env type stores common environment related items
type Env struct {
	Db     *sql.DB
	Logger *zap.Logger
}

// Init Constructs Env type to be passed around to functions
func Init() (*Env, error) {

	logger, _ := zap.NewProduction()

	// returns an open database handle of 0 or more underlying connections
	// func NewDB() (*sql.DB, error)
	sqldb, err := newDB()

	if err != nil {
		return nil, err
	}

	environment := &Env{Db: sqldb, Logger: logger}

	return environment, nil

}

// newDB returns an open database handle of 0 or more underlying connections
func newDB() (*sql.DB, error) {

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
