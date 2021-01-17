package movie_test

import (
	"reflect"
	"testing"
	"time"

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
	gotMovie, gotError := movie.NewMovie(uid, externalID, &u)
	if gotError != nil {

		if gotMovie.ID != uid {
			t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.ID, gotMovie.ID)
		}
		if gotMovie.ExternalID != wantMovie.ExternalID {
			t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.ExternalID, gotMovie.ExternalID)
		}
		if gotMovie.CreateUser != wantMovie.CreateUser {
			t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.CreateUser, gotMovie.CreateUser)
		}
		if gotMovie.UpdateUser != wantMovie.UpdateUser {
			t.Errorf("Want: %v\nGot: %v\n\n", wantMovie.UpdateUser, gotMovie.UpdateUser)
		}
	}
}

func TestSetExternalID(t *testing.T) {
	u := newValidUser()
	uid, _ := uuid.NewUUID()
	externalID := "externalID"
	externalID2 := "externalIDUpdated"

	gotMovie, _ := movie.NewMovie(uid, externalID, &u)

	gotMovie.SetExternalID(externalID2)

	if gotMovie.ExternalID != externalID2 {
		t.Errorf("Want: %v\nGot: %v\n\n", gotMovie.ExternalID, externalID2)
	}
}

func TestSetTitle(t *testing.T) {
	u := newValidUser()
	uid, _ := uuid.NewUUID()
	externalID := "externalID"
	Title := "Movie Title"

	gotMovie, _ := movie.NewMovie(uid, externalID, &u)

	gotMovie.SetTitle(Title)

	if gotMovie.Title != Title {
		t.Errorf("Want: %v\nGot: %v\n\n", gotMovie.Title, Title)
	}
}

func TestSetRated(t *testing.T) {
	u := newValidUser()
	uid, _ := uuid.NewUUID()
	externalID := "externalID"
	Rated := "R"

	gotMovie, _ := movie.NewMovie(uid, externalID, &u)

	gotMovie.SetRated(Rated)

	if gotMovie.Rated != Rated {
		t.Errorf("Want: %v\nGot: %v\n\n", gotMovie.Rated, Rated)
	}
}

func TestSetReleasedOk(t *testing.T) {
	newRealeased := time.Now()

	u := newValidUser()
	uid, _ := uuid.NewUUID()
	externalID := "externalID"

	gotMovie, _ := movie.NewMovie(uid, externalID, &u)

	gotMovie, _ = gotMovie.SetReleased(newRealeased.Format(time.RFC3339))

	if gotMovie.Released != newRealeased {
		t.Errorf("Want: %v\nGot: %v\n\n", newRealeased, gotMovie.Released)
	}

	//if e.Error() != "" {
	//t.Errorf("Error: %v", e)
	//}
}

func TestSetReleasedWrong(t *testing.T) {
	newRealeased := "wrong-time"

	u := newValidUser()
	uid, _ := uuid.NewUUID()
	externalID := "externalID"

	gotMovie, _ := movie.NewMovie(uid, externalID, &u)

	_, e := gotMovie.SetReleased(newRealeased)
	_, err := time.Parse(time.RFC3339, newRealeased)

	want := errs.E(errs.Validation,
		errs.Code("invalid_date_format"),
		errs.Parameter("release_date"),
		errors.WithStack(err))

	if e.Error() != want.Error() {
		t.Errorf("\nWant: %v\nGot: %v\n\n", want, e)
	}
}
