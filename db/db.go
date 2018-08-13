package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-API-template/errors"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
)

// name defines database name
type name int

const (
	// AppDB represents main application database
	AppDB name = iota

	// LogDB represents http logging database
	LogDB
)

// Datastore struct stores common environment related items
type Datastore struct {
	mainDB  *sql.DB
	logDB   *sql.DB
	cacheDB *redis.Pool
}

// NewDatastore initializes the datastore struct
// NOTE: I have chosen to use the same database for logging as
// my "main" app database. I'd recommend having a separate db and
// would have a separate method to start that connection pool up and
// pass it, but since this is just an example....
func NewDatastore() (*Datastore, error) {
	const op errors.Op = "db.NewDatastore"

	// Get a mainDB object (PostgreSQL)
	mdb, err := newMainDB()
	if err != nil {
		log.Error().Err(err).Msg("Error returned from newMainDB")
		return nil, err
	}

	// Get a Redis Pool from redigo client
	cDB := newCacheDb()

	// For now, store mainDB object as mainDB and logDB as they are
	// currently the same. cacheDB is Redis
	return &Datastore{mainDB: mdb, logDB: mdb, cacheDB: cDB}, nil
}

// NewMainDB returns an open database handle of 0 or more underlying connections
func newMainDB() (*sql.DB, error) {
	const op errors.Op = "db.newMainDB"

	// Get Database connection credentials from environment variables
	dbName := os.Getenv("PG_DBNAME_TEST")
	dbUser := os.Getenv("PG_USERNAME_TEST")
	dbPassword := os.Getenv("PG_PASSWORD_TEST")
	dbHost := os.Getenv("PG_HOST_TEST")
	dbPort, err := strconv.Atoi(os.Getenv("PG_PORT_TEST"))
	if err != nil {
		log.Error().Err(err).Msg("Unable to complete string to int conversion for dbPort")
		return nil, err
	}

	// Craft string for database connection
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Error().Err(err).Msg("Error returned from sql.Open")
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Error().Err(err).Msg("Error returned from db.Ping")
		return nil, err
	}
	return db, nil
}

// NewMainDB returns an pool of redis connections from
// which an application can get a new connection
func newCacheDb() *redis.Pool {
	const op errors.Op = "db.newCacheDb"
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 80,
		// max number of connections
		MaxActive: 12000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// RedisConn gets a connection from ds.cacheDB redis cache
func (ds Datastore) RedisConn() (redis.Conn, error) {
	const op errors.Op = "db.RedisConn"

	conn := ds.cacheDB.Get()

	err := conn.Err()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// BeginTx begins a *sql.Tx for the given db
func (ds Datastore) BeginTx(ctx context.Context, opts *sql.TxOptions, n name) (*sql.Tx, error) {
	const op errors.Op = "db.Datastore.BeginTx"

	switch n {
	case AppDB:
		// Calls the BeginTx method of the mainDb opened database
		mtx, err := ds.mainDB.BeginTx(ctx, opts)
		if err != nil {
			return nil, errors.E(op, err)
		}

		return mtx, nil
	case LogDB:
		// Calls the BeginTx method of the mogDB opened database
		ltx, err := ds.logDB.BeginTx(ctx, opts)
		if err != nil {
			return nil, errors.E(op, err)
		}

		return ltx, nil
	default:
		return nil, errors.E(op, "Unexpected Database Name")
	}
}

// FinalizeTx will attempt to commit or rollback the db transaction
// If the commit is successful, no error will be nil
// If the commit is not successful, a rollback will be attempted and
// the error will not be nil
func FinalizeTx(ctx context.Context, log zerolog.Logger, tx *sql.Tx, commit bool) error {

	if commit {
		err := tx.Commit()
		if err != nil {
			return err
		}
	} else {
		err := tx.Rollback()
		if err != nil {
			return err
		}
	}
	return nil
}
