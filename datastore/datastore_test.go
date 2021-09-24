package datastore

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

func TestNewPostgreSQLDSN(t *testing.T) {
	c := qt.New(t)

	got := NewPostgreSQLDSN("localhost", "go_api_basic", "postgres", "", 5432)

	want := PostgreSQLDSN{
		Host:     "localhost",
		Port:     5432,
		DBName:   "go_api_basic",
		User:     "postgres",
		Password: "",
	}

	c.Assert(got, qt.Equals, want)
}

func TestPostgreSQLDSN_String(t *testing.T) {
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
			dsn := PostgreSQLDSN{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				DBName:   tt.fields.DBName,
				User:     tt.fields.User,
				Password: tt.fields.Password,
			}
			if got := dsn.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatastore_Pool(t *testing.T) {
	c := qt.New(t)

	ctx := context.Background()
	lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

	ogpool, cleanup, err := NewPostgreSQLPool(ctx, NewPostgreSQLDSN("localhost", "go_api_basic", "postgres", "", 5432), lgr)
	t.Cleanup(cleanup)
	if err != nil {
		t.Fatal(err)
	}
	ds := Datastore{dbpool: ogpool}
	dbpool := ds.Pool()

	c.Assert(dbpool, qt.Equals, ogpool)
}

func TestNewDatastore(t *testing.T) {
	t.Run("numbers", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

		dbpool, cleanup, err := NewPostgreSQLPool(ctx, NewPostgreSQLDSN("localhost", "go_api_basic", "postgres", "", 5432), lgr)
		c.Assert(err, qt.IsNil)
		t.Cleanup(cleanup)

		got := NewDatastore(dbpool)
		want := Datastore{dbpool: dbpool}

		c.Assert(got, qt.Equals, want)
	})
}

func TestDatastore_BeginTx(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := NewPostgreSQLDSN("localhost", "go_api_basic", "postgres", "", 5432)
		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

		dbpool, cleanup, err := NewPostgreSQLPool(ctx, dsn, lgr)
		c.Assert(err, qt.IsNil)
		t.Cleanup(cleanup)

		ds := NewDatastore(dbpool)

		tx, err := ds.BeginTx(ctx)
		c.Assert(err, qt.IsNil)

		// the cleanup function pgxpool.Pool.Close() blocks until all connections have been returned to the pool
		// we have to finalize the transaction to close the pool (either commit or rollback)
		err = tx.Rollback(ctx)
		c.Assert(err, qt.IsNil)
	})

	t.Run("closed pool", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		dsn := NewPostgreSQLDSN("localhost", "go_api_basic", "postgres", "", 5432)
		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

		dbpool, cleanup, err := NewPostgreSQLPool(ctx, dsn, lgr)
		c.Assert(err, qt.IsNil)
		// cleanup closes the pool
		cleanup()

		ds := NewDatastore(dbpool)

		_, err = ds.BeginTx(ctx)
		c.Assert(errors.Is(err, puddle.ErrClosedPool), qt.IsTrue)
	})

	t.Run("nil pool", func(t *testing.T) {
		c := qt.New(t)

		ds := NewDatastore(nil)
		ds.dbpool = nil

		ctx := context.Background()
		_, err := ds.BeginTx(ctx)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Database, "db pool cannot be nil"))
	})

}

func TestDatastore_RollbackTx(t *testing.T) {
	type fields struct {
		dbpool *pgxpool.Pool
	}
	type args struct {
		ctx context.Context
		tx  pgx.Tx
		err error
	}

	ctx := context.Background()
	dsn := NewPostgreSQLDSN("localhost", "go_api_basic", "postgres", "", 5432)
	lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

	dbpool, cleanup, err := NewPostgreSQLPool(ctx, dsn, lgr)
	t.Cleanup(cleanup)
	if err != nil {
		t.Errorf("datastore.NewPostgreSQLDB error = %v", err)
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

	err = errs.E("some error")

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"typical", fields{dbpool}, args{ctx, tx, err}},
		{"nil tx", fields{dbpool}, args{ctx, nil, err}},
		{"already committed tx", fields{dbpool}, args{ctx, tx2, err}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			ds := &Datastore{
				dbpool: tt.fields.dbpool,
			}
			rollbackErr := ds.RollbackTx(tt.args.ctx, tt.args.tx, tt.args.err)
			// I'm sending an *errs.Error, so RollbackTx will only return an *errs.Error
			e, _ := rollbackErr.(*errs.Error)
			t.Logf("error = %v", e)
			if tt.args.tx == nil && e.Code != "nil_tx" {
				t.Fatalf("ds.RollbackTx() tx was nil, but incorrect error returned = %v", e)
			}
			// I know this is weird, but it's the only way I could think to test this.
			if tt.name == "already committed tx" && e.Code != "rollback_err" {
				t.Fatalf("ds.RollbackTx() tx was already committed, but incorrect error returned = %v", e)
			}
		})
	}
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
	dsn := NewPostgreSQLDSN("localhost", "go_api_basic", "postgres", "", 5432)
	lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

	dbpool, cleanup, err := NewPostgreSQLPool(ctx, dsn, lgr)
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
			ds := &Datastore{
				dbpool: tt.fields.dbpool,
			}
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
			got := NewNullString(tt.args.s)
			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

func TestNewNullInt64(t *testing.T) {
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
			if got := NewNullInt64(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
