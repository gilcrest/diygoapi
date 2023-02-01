package sqldb_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/puddle"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/logger"
	"github.com/gilcrest/diygoapi/sqldb"
)

func TestPostgreSQLDSN_ConnectionKeywordValueString(t *testing.T) {
	type fields struct {
		Host     string
		Port     int
		DBName   string
		User     string
		Password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"with password", fields{Host: "localhost", Port: 8080, DBName: "go_api_basic", User: "postgres", Password: "supahsecret"}, "host=localhost port=8080 dbname=go_api_basic user=postgres password=supahsecret sslmode=disable"},
		{"without password", fields{Host: "localhost", Port: 8080, DBName: "go_api_basic", User: "postgres", Password: ""}, "host=localhost port=8080 dbname=go_api_basic user=postgres sslmode=disable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := sqldb.PostgreSQLDSN{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				DBName:   tt.fields.DBName,
				User:     tt.fields.User,
				Password: tt.fields.Password,
			}
			if got := dsn.KeywordValueConnectionString(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

//func TestDB_Pool(t *testing.T) {
//	c := qt.New(t)
//
//	ctx := context.Background()
//	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)
//
//	dsn := newPostgreSQLDSN(t)
//
//	ogpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
//	t.Cleanup(cleanup)
//	if err != nil {
//		t.Fatal(err)
//	}
//	ds := sqldb.NewDB(ogpool)
//	dbpool := ds.Pool()
//
//	c.Assert(dbpool, qt.Equals, ogpool)
//}

func TestDB_BeginTx(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := newPostgreSQLDSN(t)
		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
		c.Assert(err, qt.IsNil)
		t.Cleanup(cleanup)

		ds := sqldb.NewDB(dbpool)

		var tx pgx.Tx
		tx, err = ds.BeginTx(ctx)
		c.Assert(err, qt.IsNil)

		// the cleanup function pgxpool.Pool.Close() blocks until all connections have been returned to the pool
		// we have to finalize the transaction to close the pool (either commit or rollback)
		err = tx.Rollback(ctx)
		c.Assert(err, qt.IsNil)
	})
	t.Run("closed pool", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := newPostgreSQLDSN(t)
		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
		c.Assert(err, qt.IsNil)
		// cleanup closes the pool
		cleanup()

		ds := sqldb.NewDB(dbpool)

		_, err = ds.BeginTx(ctx)
		c.Assert(errors.Is(err, puddle.ErrClosedPool), qt.IsTrue)
	})

	t.Run("nil pool", func(t *testing.T) {
		c := qt.New(t)

		var err error

		ds := sqldb.NewDB(nil)

		ctx := context.Background()
		_, err = ds.BeginTx(ctx)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Database, "db pool cannot be nil"))
	})

}

func TestDatastore_RollbackTx(t *testing.T) {
	t.Run("rollback due to error", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := newPostgreSQLDSN(t)
		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		// get a *pgxpool.Pool and setup datastore.Datastore
		dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
		t.Cleanup(cleanup)
		if err != nil {
			t.Errorf("datastore.NewPostgreSQLDB error = %v", err)
		}
		ds := sqldb.NewDB(dbpool)

		// begin a new tx from pool in datastore.Datastore
		var tx pgx.Tx
		tx, err = ds.BeginTx(ctx)
		if err != nil {
			t.Errorf("ds.BeginTx error = %v", err)
		}

		// create fake error to mimic an error that might trigger a rollback
		fakeErr := errs.E(errs.Validation, errs.Code("INVALID_TOKEN"), "some validation error")

		// attempt to roll back the transaction. The original error (fakeErr)
		// should be returned as the response
		rollbackErr := ds.RollbackTx(ctx, tx, fakeErr)

		// assert that it does indeed match fakeErr and nothing else has been added
		var errsErr *errs.Error
		if errors.As(rollbackErr, &errsErr) {
			c.Assert(errsErr.Kind, qt.Equals, errs.Validation)
			c.Assert(errsErr.Code, qt.DeepEquals, errs.Code("INVALID_TOKEN"))
			c.Assert(errsErr.Error(), qt.Equals, fakeErr.Error())
		} else {
			c.Fatalf("rollbackErr is invalid: %v", rollbackErr)
		}
	})
	t.Run("nil tx", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := newPostgreSQLDSN(t)
		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		// get a *pgxpool.Pool and setup datastore.Datastore
		dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
		t.Cleanup(cleanup)
		if err != nil {
			t.Errorf("datastore.NewPostgreSQLDB error = %v", err)
		}
		ds := sqldb.NewDB(dbpool)

		// create fake error to mimic an error that might trigger a rollback
		fakeErr := errs.E(errs.Validation, errs.Code("INVALID_TOKEN"), "some validation error")

		// attempt to roll back the transaction. The original error (fakeErr)
		// should be returned combined with the error created in RollbackTX
		rollbackErr := ds.RollbackTx(ctx, nil, fakeErr)

		// assert that the combined error string matches what is expected
		// as well as ensure the Kind and Code are correct
		var errsErr *errs.Error
		if errors.As(rollbackErr, &errsErr) {
			c.Assert(errsErr.Kind, qt.Equals, errs.Database)
			c.Assert(errsErr.Code, qt.DeepEquals, errs.Code("nil_tx"))
			c.Assert(errsErr.Error(), qt.Equals, fmt.Sprintf("RollbackTx() error = tx cannot be nil: Original error = %s", fakeErr.Error()))
		} else {
			c.Fatalf("rollbackErr is invalid: %v", rollbackErr)
		}
	})
	t.Run("already committed tx with error", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := newPostgreSQLDSN(t)
		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		// get a *pgxpool.Pool and setup datastore.Datastore
		dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
		t.Cleanup(cleanup)
		if err != nil {
			t.Errorf("datastore.NewPostgreSQLDB error = %v", err)
		}
		ds := sqldb.NewDB(dbpool)

		// begin a new tx from pool in datastore.Datastore
		var tx pgx.Tx
		tx, err = ds.BeginTx(ctx)
		if err != nil {
			t.Errorf("ds.BeginTx error = %v", err)
		}
		// commit the tx - rollback should not work
		err = tx.Commit(ctx)
		if err != nil {
			t.Errorf("tx.Commit error = %v", err)
		}

		// create fake error to mimic an error that might trigger a rollback
		fakeErr := errs.E(errs.Validation, errs.Code("INVALID_TOKEN"), "some validation error")

		// attempt to roll back the transaction. The original error (fakeErr)
		// should be returned combined with the error created in RollbackTX
		rollbackErr := ds.RollbackTx(ctx, tx, fakeErr)

		// assert that the combined error string matches what is expected
		// as well as ensure the Kind and Code are correct
		var errsErr *errs.Error
		if errors.As(rollbackErr, &errsErr) {
			c.Assert(errsErr.Kind, qt.Equals, errs.Validation)
			c.Assert(errsErr.Code, qt.DeepEquals, errs.Code("INVALID_TOKEN"))
			c.Assert(errsErr.Error(), qt.Equals, fakeErr.Error())
		} else {
			c.Fatalf("rollbackErr is invalid: %v", rollbackErr)
		}
	})
	t.Run("already committed tx with no error", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := newPostgreSQLDSN(t)
		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		// get a *pgxpool.Pool and setup datastore.Datastore
		dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
		t.Cleanup(cleanup)
		if err != nil {
			t.Errorf("datastore.NewPostgreSQLDB error = %v", err)
		}
		ds := sqldb.NewDB(dbpool)

		// begin a new tx from pool in datastore.Datastore
		var tx pgx.Tx
		tx, err = ds.BeginTx(ctx)
		if err != nil {
			t.Errorf("ds.BeginTx error = %v", err)
		}
		// commit the tx - rollback should not work
		err = tx.Commit(ctx)
		if err != nil {
			t.Errorf("tx.Commit error = %v", err)
		}

		// attempt to roll back the already committed transaction.
		// There is no error, so this is the typical case, the tx
		// should be closed and the error should be returned as nil
		rollbackErr := ds.RollbackTx(ctx, tx, nil)

		c.Assert(rollbackErr, qt.IsNil)
	})
	t.Run("deferred rollback (with error)", func(t *testing.T) {
		c := qt.New(t)

		deferErr := checkDefer(t)
		// assert that the error string matches what is expected
		// as well as ensure the Kind and Code are correct
		var errsErr *errs.Error
		if errors.As(deferErr, &errsErr) {
			c.Assert(errsErr.Kind, qt.Equals, errs.Validation)
			c.Assert(errsErr.Code, qt.DeepEquals, errs.Code("DATA_VALIDATION"))
			c.Assert(errsErr.Error(), qt.Equals, "This validation happened.")
		} else {
			c.Fatalf("deferErr is invalid: %v", deferErr)
		}
	})
	t.Run("deferred rollback (nil error)", func(t *testing.T) {
		c := qt.New(t)

		_, deferErr := checkDefer2FieldsNilError(t)
		c.Assert(deferErr, qt.IsNil)
	})
	t.Run("deferred rollback (nil tx)", func(t *testing.T) {
		c := qt.New(t)

		_, deferErr := checkDefer2FieldsNilTx(t)
		// assert that the combined error string matches what is expected
		// as well as ensure the Kind and Code are correct
		var errsErr *errs.Error
		if errors.As(deferErr, &errsErr) {
			c.Assert(errsErr.Kind, qt.Equals, errs.Database)
			c.Assert(errsErr.Code, qt.DeepEquals, errs.Code("nil_tx"))
			c.Assert(errsErr.Error(), qt.Equals, "RollbackTx() error = tx cannot be nil: Original error is nil")
		} else {
			c.Fatalf("rollbackErr is invalid: %v", deferErr)
		}
	})
	// take down db during this test in order to put tx in bad state
	// this test will typically be skipped, just doing it to see how it
	// works
	t.Run("deferred rollback (kill db)", func(t *testing.T) {
		t.Skip()

		c := qt.New(t)

		deferErr := checkDeferKillDB(t)
		// assert that the combined error string matches what is expected
		// as well as ensure the Kind and Code are correct
		var errsErr *errs.Error
		if errors.As(deferErr, &errsErr) {
			c.Assert(errsErr.Kind, qt.Equals, errs.Database)
			c.Assert(errsErr.Code, qt.DeepEquals, errs.Code("rollback_err"))
			c.Assert(errsErr.Error(), qt.Equals, "PG Error Code: 57P01, PG Error Message: terminating connection due to administrator command, RollbackTx() error = FATAL: terminating connection due to administrator command (SQLSTATE 57P01): Original error = This validation happened.")
		} else {
			c.Fatalf("rollbackErr is invalid: %v", deferErr)
		}
	})
}

func checkDefer(t *testing.T) (err error) {
	const op errs.Op = "sqldb_test/checkDefer"

	ctx := context.Background()
	dsn := newPostgreSQLDSN(t)
	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

	// get a *pgxpool.Pool and setup datastore.Datastore
	var (
		dbpool  *pgxpool.Pool
		cleanup func()
	)
	dbpool, cleanup, err = sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
	t.Cleanup(cleanup)
	if err != nil {
		t.Fatalf("datastore.NewPostgreSQLDB error = %v", err)
	}
	ds := sqldb.NewDB(dbpool)

	// begin a new tx from pool in datastore.Datastore
	var tx pgx.Tx
	tx, err = ds.BeginTx(ctx)
	if err != nil {
		t.Fatalf("ds.BeginTx error = %v", err)
	}
	defer func() {
		err = ds.RollbackTx(ctx, tx, err)
	}()

	err = errs.E(op, errs.Validation, errs.Code("DATA_VALIDATION"), "This validation happened.")
	if err != nil {
		return err
	}

	// commit the tx
	err = ds.CommitTx(ctx, tx)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

type fakeUser struct {
	username string
}

func checkDefer2FieldsNilError(t *testing.T) (fu fakeUser, err error) {
	const op errs.Op = "sqldb_test/checkDefer2FieldsNilError"

	ctx := context.Background()
	dsn := newPostgreSQLDSN(t)
	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

	// get a *pgxpool.Pool and setup datastore.Datastore
	var (
		dbpool  *pgxpool.Pool
		cleanup func()
	)
	dbpool, cleanup, err = sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
	t.Cleanup(cleanup)
	if err != nil {
		t.Fatalf("datastore.NewPostgreSQLDB error = %v", err)
	}
	ds := sqldb.NewDB(dbpool)

	// begin a new tx from pool in datastore.Datastore
	var tx pgx.Tx
	tx, err = ds.BeginTx(ctx)
	if err != nil {
		t.Fatalf("ds.BeginTx error = %v", err)
	}
	defer func() {
		err = ds.RollbackTx(ctx, tx, err)
	}()

	// commit the tx
	err = ds.CommitTx(ctx, tx)
	if err != nil {
		return fakeUser{}, errs.E(op, err)
	}

	return fakeUser{}, nil
}

func checkDefer2FieldsNilTx(t *testing.T) (fu fakeUser, err error) {
	const op errs.Op = "sqldb_test/checkDefer2FieldsNilTx"

	ctx := context.Background()
	dsn := newPostgreSQLDSN(t)
	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

	// get a *pgxpool.Pool and setup datastore.Datastore
	var (
		dbpool  *pgxpool.Pool
		cleanup func()
	)
	dbpool, cleanup, err = sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
	t.Cleanup(cleanup)
	if err != nil {
		t.Fatalf("datastore.NewPostgreSQLDB error = %v", err)
	}
	ds := sqldb.NewDB(dbpool)

	// begin a new tx from pool in datastore.Datastore
	var tx pgx.Tx
	tx, err = ds.BeginTx(ctx)
	if err != nil {
		t.Errorf("ds.BeginTx error = %v", err)
	}
	defer func() {
		err = ds.RollbackTx(ctx, nil, err)
	}()

	// commit the tx
	err = ds.CommitTx(ctx, tx)
	if err != nil {
		return fakeUser{}, errs.E(op, err)
	}

	return fakeUser{}, err
}

func checkDeferKillDB(t *testing.T) (err error) {
	const op errs.Op = "sqldb_test/checkDeferKillDB"

	ctx := context.Background()
	dsn := newPostgreSQLDSN(t)
	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

	// get a *pgxpool.Pool and setup datastore.Datastore
	dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
	t.Cleanup(cleanup)
	if err != nil {
		t.Fatalf("datastore.NewPostgreSQLDB error = %v", err)
	}
	ds := sqldb.NewDB(dbpool)

	// begin a new tx from pool in datastore.Datastore
	var tx pgx.Tx
	tx, err = ds.BeginTx(ctx)
	if err != nil {
		t.Fatalf("ds.BeginTx error = %v", err)
	}
	defer func() {
		err = ds.RollbackTx(ctx, tx, err)
	}()

	t.Log("Go shutdown database server immediately for this test to work")
	time.Sleep(30 * time.Second)

	return errs.E(op, errs.Validation, errs.Code("DATA_VALIDATION"), "This validation happened.")
}

func TestDatastore_CommitTx(t *testing.T) {
	type fields struct {
		dbpool *pgxpool.Pool
	}
	type args struct {
		ctx context.Context
		tx  pgx.Tx
	}

	ctx := context.Background()
	dsn := newPostgreSQLDSN(t)
	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

	dbpool, cleanup, err := sqldb.NewPostgreSQLPool(ctx, lgr, dsn)
	t.Cleanup(cleanup)
	if err != nil {
		t.Errorf("datastore.NewPostgreSQLPool error = %v", err)
	}
	tx, err := dbpool.Begin(ctx)
	if err != nil {
		t.Errorf("dbpool.Begin error = %v", err)
	}
	tx2, err := dbpool.Begin(ctx)
	if err != nil {
		t.Errorf("dbpool.Begin error = %v", err)
	}
	err = tx2.Commit(ctx)
	if err != nil {
		t.Errorf("tx2.Commit() error = %v", err)
	}
	tx3, err := dbpool.Begin(ctx)
	if err != nil {
		t.Errorf("dbpool.Begin error = %v", err)
	}
	err = tx3.Rollback(ctx)
	if err != nil {
		t.Errorf("tx2.Commit() error = %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{dbpool}, args{ctx, tx}, false},
		{"already committed", fields{dbpool}, args{ctx, tx2}, true},
		{"already rolled back", fields{dbpool}, args{ctx, tx3}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := sqldb.NewDB(tt.fields.dbpool)
			if commitErr := ds.CommitTx(tt.args.ctx, tt.args.tx); (commitErr != nil) != tt.wantErr {
				t.Errorf("CommitTx() error = %v, wantErr %v", commitErr, tt.wantErr)
			}
		})
	}
}

func TestNewNullString(t *testing.T) {
	c := qt.New(t)
	type args struct {
		s string
	}

	wantNotNull := sql.NullString{String: "not null", Valid: true}
	wantNull := sql.NullString{String: "", Valid: false}
	tests := []struct {
		name string
		args args
		want sql.NullString
	}{
		{"not null string", args{s: "not null"}, wantNotNull},
		{"null string", args{s: ""}, wantNull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diygoapi.NewNullString(tt.args.s)
			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

func TestNewNullInt64(t *testing.T) {
	c := qt.New(t)

	type args struct {
		i int64
	}
	tests := []struct {
		name string
		args args
		want sql.NullInt64
	}{
		{"has value", args{i: 23}, sql.NullInt64{Int64: 23, Valid: true}},
		{"zero value", args{i: 0}, sql.NullInt64{Int64: 0, Valid: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diygoapi.NewNullInt64(tt.args.i)
			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

// newPGDatasourceName is a test helper to get a PGDatasourceName
// from environment variables
func newPostgreSQLDSN(t *testing.T) sqldb.PostgreSQLDSN {
	t.Helper()

	var (
		dbHost       string
		dbPort       int
		dbName       string
		dbUser       string
		dbPassword   string
		dbSearchPath string
		ok           bool
		err          error
	)

	dbHost, ok = os.LookupEnv(sqldb.DBHostEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBHostEnv)
	}

	var p string
	p, ok = os.LookupEnv(sqldb.DBPortEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBPortEnv)
	}
	dbPort, err = strconv.Atoi(p)
	if err != nil {
		t.Fatalf("Unable to convert db port %s to int", p)
	}

	dbName, ok = os.LookupEnv(sqldb.DBNameEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBNameEnv)
	}

	dbUser, ok = os.LookupEnv(sqldb.DBUserEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBUserEnv)
	}

	dbPassword, ok = os.LookupEnv(sqldb.DBPasswordEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBPasswordEnv)
	}

	dbSearchPath, ok = os.LookupEnv(sqldb.DBSearchPathEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBSearchPathEnv)
	}

	return sqldb.PostgreSQLDSN{
		Host:       dbHost,
		Port:       dbPort,
		DBName:     dbName,
		SearchPath: dbSearchPath,
		User:       dbUser,
		Password:   dbPassword,
	}
}
