package movie_test

import (
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Returns a valid User with mocked data
func newValidUser() user.User {
	return user.User{
		Email:        "foo@bar.com",
		LastName:     "Bar",
		FirstName:    "Foo",
		FullName:     "Foo Bar",
		HostedDomain: "example.com",
		PictureURL:   "example.com/profile.png",
		ProfileLink:  "example.com/FooBar",
	}
}

// Returns an invalid user defined by the method user.IsValid()
func newInvalidUser() user.User {
	return user.User{
		Email:        "",
		LastName:     "",
		FirstName:    "",
		FullName:     "",
		HostedDomain: "example.com",
		PictureURL:   "example.com/profile.png",
		ProfileLink:  "example.com/FooBar",
	}
}

// Testing error when sent a nil uuid
func TestNewMovieErrorUuid(t *testing.T) {
	t.Helper()

	u := newValidUser()
	wantError := errs.E(errs.Validation, errs.Parameter("ID"), errors.New(errs.MissingField("ID").Error()))
	if gotMovie, gotError := movie.NewMovie(uuid.UUID{}, "randomExternalId", &u); !reflect.DeepEqual(wantError.Error(), gotError.Error()) && gotMovie != nil {
		t.Errorf("Want: %v\nGot: %v", wantError, gotError)
	}
}

// Testing error when sent a nil ExtlID
func TestNewMovieErrorExtlID(t *testing.T) {
	t.Helper()

	u := newValidUser()
	uid, _ := uuid.NewUUID()
	wantError := errs.E(errs.Validation, errs.Parameter("ID"), errors.New(errs.MissingField("ID").Error()))
	if gotMovie, gotError := movie.NewMovie(uid, "", &u); !reflect.DeepEqual(wantError.Error(), gotError.Error()) && gotMovie != nil {
		t.Errorf("Want: %v\nGot: %v", wantError, gotError)
	}
}

// Testing error when User invalid
func TestNewMovieErrorInvalidUser(t *testing.T) {
	t.Helper()

	u := newInvalidUser()
	uid, _ := uuid.NewUUID()

	wantError := errs.E(errs.Validation, errs.Parameter("User"), errors.New("User is invalid"))

	if gotMovie, gotError := movie.NewMovie(uid, "externalID", &u); !reflect.DeepEqual(wantError.Error(), gotError.Error()) && gotMovie != nil {
		t.Errorf("Want: %v\nGot: %v", wantError, gotError)
	}
}

// Testing creating NewMovie
func TestNewMovie(t *testing.T) {
	t.Helper()

	u := newValidUser()
	uid, _ := uuid.NewUUID()
	externalID := "externalID"

	wantMovie := movie.Movie{
		ID:         uid,
		ExternalID: externalID,
		CreateUser: u,
		UpdateUser: u,
	}

	if gotMovie, gotError := movie.NewMovie(uid, externalID, &u); (gotMovie.ID != uid ||
		gotMovie.ExternalID != wantMovie.ExternalID ||
		gotMovie.CreateUser != wantMovie.CreateUser ||
		gotMovie.UpdateUser != wantMovie.UpdateUser) && gotError == nil {
		t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.ID, gotMovie.ID)
		t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.ExternalID, gotMovie.ExternalID)
		t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.CreateUser, gotMovie.CreateUser)
		t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.UpdateUser, gotMovie.UpdateUser)
	}
}
