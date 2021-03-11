package moviestore

import (
	"context"
	"testing"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/gilcrest/go-api-basic/domain/user/usertest"
	"github.com/google/uuid"
)

// NewMovieDBHelper creates/inserts a new movie in the db and optionally
// registers a t.Cleanup function to delete it. The insert and
// delete are both in separate database transactions
func NewMovieDBHelper(t *testing.T, ctx context.Context, ds datastore.Datastorer) (m *movie.Movie, cleanup func()) {
	t.Helper()

	m = newMovie(t)

	defaultTransactor := NewDefaultTransactor(ds)

	err := defaultTransactor.Create(ctx, m)
	if err != nil {
		t.Fatalf("defaultTransactor.Create error = %v", err)
	}

	cleanup = func() {
		err := defaultTransactor.Delete(ctx, m)
		if err != nil {
			t.Fatalf("t.Cleanup defaultTransactor.Delete error = %v", err)
		}
	}

	return m, cleanup
}

func newMovie(t *testing.T) *movie.Movie {
	t.Helper()

	id := uuid.New()
	rsg := random.DefaultStringGenerator{}
	extlID, err := rsg.CryptoString(15)
	if err != nil {
		t.Fatalf("random.CryptoString() error = %v", err)
	}
	u := usertest.NewUser(t)
	m, err := movie.NewMovie(id, extlID, u)
	if err != nil {
		t.Fatalf("movie.NewMovie() error = %v", err)
	}
	return m
}
