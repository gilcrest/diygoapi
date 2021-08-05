package moviestore

import (
	"context"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/gilcrest/go-api-basic/domain/user/usertest"
	"github.com/google/uuid"
)

// NewMovieDBHelper creates/inserts a new movie in the db and optionally
// registers a t.Cleanup function to delete it. The insert and
// delete are both in separate database transactions
func NewMovieDBHelper(ctx context.Context, t *testing.T, ds Datastorer) (m *movie.Movie, cleanup func()) {
	t.Helper()

	m = newMovie(t)

	movieTransactor := NewTransactor(ds)

	err := movieTransactor.Create(ctx, m)
	if err != nil {
		t.Fatalf("movieTransactor.Create error = %v", err)
	}

	cleanup = func() {
		err := movieTransactor.Delete(ctx, m)
		if err != nil {
			t.Fatalf("t.Cleanup movieTransactor.Delete error = %v", err)
		}
	}

	return m, cleanup
}

func newMovie(t *testing.T) *movie.Movie {
	t.Helper()

	id := uuid.New()
	rsg := random.StringGenerator{}
	extlID, err := rsg.CryptoString(15)
	if err != nil {
		t.Fatalf("random.CryptoString() error = %v", err)
	}
	u := usertest.NewUser(t)
	m, err := movie.NewMovie(id, extlID, u)
	if err != nil {
		t.Fatalf("movie.NewMovie() error = %v", err)
	}
	m, _ = m.SetReleased("1984-03-02T00:00:00Z")
	m.SetTitle("Repo Man").
		SetRated("R").
		SetRunTime(92).
		SetWriter("Alex Cox").
		SetDirector("Alex Cox")

	return m
}
