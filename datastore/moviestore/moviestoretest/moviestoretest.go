// Package moviestoretest provides testing helper functions for the
// moviestore package
package moviestoretest

import (
	"context"
	"testing"
	"time"

	"github.com/gilcrest/go-api-basic/domain/user/usertest"

	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/google/uuid"
)

// NewMockTransactor is an initializer for MockTransactor
func NewMockTransactor(t *testing.T) MockTransactor {
	return MockTransactor{t: t}
}

// MockTransactor is a mock which satisfies the moviestore.Transactor
// interface
type MockTransactor struct {
	t *testing.T
}

func (mt MockTransactor) Create(ctx context.Context, m *movie.Movie) error {
	return nil
}

func (mt MockTransactor) Update(ctx context.Context, m *movie.Movie) error {
	return nil
}

func (mt MockTransactor) Delete(ctx context.Context, m *movie.Movie) error {
	return nil
}

// NewMockSelector is an initializer for MockSelector
func NewMockSelector(t *testing.T) MockSelector {
	return MockSelector{t: t}
}

// MockSelector is a mock which satisfies the moviestore.Selector
// interface
type MockSelector struct {
	t *testing.T
}

// FindByID mocks finding a movie by External ID
func (ms MockSelector) FindByID(ctx context.Context, s string) (*movie.Movie, error) {

	// get test user
	u := usertest.NewUser(ms.t)

	// mock create/update timestamp
	cuTime := time.Date(2008, 1, 8, 06, 54, 0, 0, time.UTC)

	return &movie.Movie{
		ID:         uuid.MustParse("f118f4bb-b345-4517-b463-f237630b1a07"),
		ExternalID: "kCBqDtyAkZIfdWjRDXQG",
		Title:      "Repo Man",
		Rated:      "R",
		Released:   time.Date(1984, 3, 2, 0, 0, 0, 0, time.UTC),
		RunTime:    92,
		Director:   "Alex Cox",
		Writer:     "Alex Cox",
		CreateUser: u,
		CreateTime: cuTime,
		UpdateUser: u,
		UpdateTime: cuTime,
	}, nil
}

// FindAll mocks finding multiple movies by External ID
func (ms MockSelector) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	// get test user
	u := usertest.NewUser(ms.t)

	// mock create/update timestamp
	cuTime := time.Date(2008, 1, 8, 06, 54, 0, 0, time.UTC)

	m1 := &movie.Movie{
		ID:         uuid.MustParse("f118f4bb-b345-4517-b463-f237630b1a07"),
		ExternalID: "kCBqDtyAkZIfdWjRDXQG",
		Title:      "Repo Man",
		Rated:      "R",
		Released:   time.Date(1984, 3, 2, 0, 0, 0, 0, time.UTC),
		RunTime:    92,
		Director:   "Alex Cox",
		Writer:     "Alex Cox",
		CreateUser: u,
		CreateTime: cuTime,
		UpdateUser: u,
		UpdateTime: cuTime,
	}

	m2 := &movie.Movie{
		ID:         uuid.MustParse("e883ebbb-c021-423b-954a-e94edb8b85b8"),
		ExternalID: "RWn8zcaTA1gk3ybrBdQV",
		Title:      "The Return of the Living Dead",
		Rated:      "R",
		Released:   time.Date(1985, 8, 16, 0, 0, 0, 0, time.UTC),
		RunTime:    91,
		Director:   "Dan O'Bannon",
		Writer:     "Russell Streiner",
		CreateUser: u,
		CreateTime: cuTime,
		UpdateUser: u,
		UpdateTime: cuTime,
	}

	return []*movie.Movie{m1, m2}, nil
}
