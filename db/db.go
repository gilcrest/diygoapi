package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

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
	mainTx  *sql.Tx
	logDB   *sql.DB
	logTx   *sql.Tx
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

// SetTx opens a database connection and starts a database transaction using the
// BeginTx method which allows for rollback of the transaction if the context
// is cancelled
func (ds *Datastore) SetTx(ctx context.Context, opts *sql.TxOptions) error {
	const op errors.Op = "db.Datastore.SetTx"

	// Calls the BeginTx method of the MainDb opened database
	mtx, err := ds.mainDB.BeginTx(ctx, opts)
	if err != nil {
		return errors.E(op, "Error returned from mainDb.BeginTx")
	}

	ds.mainTx = mtx

	// Calls the BeginTx method of the MainDb opened database
	ltx, err := ds.logDB.BeginTx(ctx, opts)
	if err != nil {
		return errors.E(op, "Error returned from logDb.BeginTx")
	}

	ds.logTx = ltx

	return nil
}

// Tx is a getter for the *sql.Tx for the given db
func (ds Datastore) Tx(n name) (*sql.Tx, error) {
	const op errors.Op = "db.Datastore.Tx"

	switch n {
	case AppDB:
		tx, err := ds.txMain()
		if err != nil {
			return nil, errors.E(op, err)
		}
		return tx, nil
	case LogDB:
		tx, err := ds.txLog()
		if err != nil {
			return nil, errors.E(op, err)
		}
		return tx, nil
	default:
		return nil, errors.E(op, "Unexpected Database Name")
	}
}

// TxMain is a getter for the mainDb *sql.Tx
func (ds Datastore) txMain() (*sql.Tx, error) {
	const op errors.Op = "db.Datastore.txMain"
	if ds.mainTx == nil {
		return nil, errors.E(op, `*sql.Tx is not initialized for mainDb. Be sure to use SetTx to start database txns `)
	}
	return ds.mainTx, nil
}

// TxLog is a getter for the mainDb *sql.Tx
func (ds Datastore) txLog() (*sql.Tx, error) {
	const op errors.Op = "db.Datastore.txLog"
	if ds.logTx == nil {
		return nil, errors.E(op, "*sql.Tx is not initialized for logDb")
	}
	return ds.logTx, nil
}

// Commit will commit the txn for the given db
func (ds Datastore) Commit(args ...name) error {
	const op errors.Op = "db.Datastore.Commit"

	if len(args) == 0 {
		errors.E(op, "call to Commit with no arguments")
	}

	for _, arg := range args {
		switch arg {
		case AppDB:
			err := ds.commitMainTx()
			if err != nil {
				return errors.E(op, err)
			}
		case LogDB:
			err := ds.commitLogTx()
			if err != nil {
				return errors.E(op, err)
			}
		default:
			return errors.E(op, "Unexpected Type")
		}
	}

	return nil
}

func (ds Datastore) commitMainTx() error {
	const op errors.Op = "db.Datastore.CommitMainTx"

	tx, err := ds.txMain()
	if err != nil {
		return errors.E(op, err)
	}
	err = tx.Commit()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (ds Datastore) commitLogTx() error {
	const op errors.Op = "db.Datastore.commitLogTx"

	tx, err := ds.txLog()
	if err != nil {
		return errors.E(op, err)
	}
	err = tx.Commit()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

// Rollback will rollback the txn for the given db
func (ds Datastore) Rollback(args ...name) error {
	const op errors.Op = "db.Datastore.Rollback"

	if len(args) == 0 {
		errors.E(op, "call to Rollback with no arguments")
	}

	for _, arg := range args {
		switch arg {
		case AppDB:
			err := ds.rollbackMainTx()
			if err != nil {
				return errors.E(op, err)
			}
		case LogDB:
			err := ds.rollbackLogTx()
			if err != nil {
				return errors.E(op, err)
			}
		default:
			return errors.E(op, "Unexpected Type")
		}
	}

	return nil
}

func (ds Datastore) rollbackMainTx() error {
	const op errors.Op = "db.Datastore.rollbackMainTx"

	tx, err := ds.txMain()
	if err != nil {
		return errors.E(op, err)
	}
	err = tx.Rollback()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (ds Datastore) rollbackLogTx() error {
	const op errors.Op = "db.Datastore.rollbackLogTx"

	tx, err := ds.txLog()
	if err != nil {
		return errors.E(op, err)
	}
	err = tx.Rollback()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
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
