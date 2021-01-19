package movie_test

import (
	qt "github.com/frankban/quicktest"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"reflect"
	"testing"
	"time"
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

func NewValidMovie() *movie.Movie {
	uid := uuid.New()
	externalID := "ExternalID"

	u := newValidUser()

	m, _ := movie.NewMovie(uid, externalID, &u)

	return m
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
	gotMovie := NewValidMovie()
	externalID2 := "externalIDUpdated"

	gotMovie.SetExternalID(externalID2)

	if gotMovie.ExternalID != externalID2 {
		t.Errorf("Want: %v\nGot: %v\n\n", gotMovie.ExternalID, externalID2)
	}
}

func TestSetTitle(t *testing.T) {
	gotMovie := NewValidMovie()
	Title := "Movie Title"

	gotMovie.SetTitle(Title)

	if gotMovie.Title != Title {
		t.Errorf("Want: %v\nGot: %v\n\n", gotMovie.Title, Title)
	}
}

func TestSetRated(t *testing.T) {
	gotMovie := NewValidMovie()
	Rated := "R"

	gotMovie.SetRated(Rated)

	if gotMovie.Rated != Rated {
		t.Errorf("Want: %v\nGot: %v\n\n", gotMovie.Rated, Rated)
	}
}

func TestSetReleasedOk(t *testing.T) {
	newReleased := "1984-01-02T15:04:05Z"
	r, err := time.Parse(time.RFC3339, newReleased)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}

	gotMovie := NewValidMovie()

	gotMovie, _ = gotMovie.SetReleased(newReleased)

	if gotMovie.Released != r {
		t.Errorf("Want: %v\nGot: %v\n\n", newReleased, gotMovie.Released)
	}
}

func TestSetReleasedWrong(t *testing.T) {
	newRealeased := "wrong-time"

	gotMovie := NewValidMovie()

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

func TestSetRunTime(t *testing.T) {
	rt := 1999

	gotMovie := NewValidMovie()

	gotMovie.SetRunTime(rt)

	if gotMovie.RunTime != rt {
		t.Errorf("\nWant: %v\nGot: %v\n\n", rt, gotMovie.RunTime)
	}
}

func TestSetDirector(t *testing.T) {
	d := "Director Drach"

	gotMovie := NewValidMovie()

	gotMovie.SetDirector(d)

	if gotMovie.Director != d {
		t.Errorf("\nWant: %v\nGot: %v\n\n", d, gotMovie.Director)
	}
}

func TestSetWriter(t *testing.T) {
	w := "Writer Drach"

	gotMovie := NewValidMovie()

	gotMovie.SetWriter(w)

	if gotMovie.Writer != w {
		t.Errorf("\nWant: %v\nGot: %v\n\n", w, gotMovie.Writer)
	}
}

func TestSetUpdateUser(t *testing.T) {
	gotMovie := NewValidMovie()

	newUser := user.User{
		Email:        "foo2@bar.com",
		LastName:     "Barw",
		FirstName:    "Foow",
		FullName:     "Foow Barw",
		HostedDomain: "example.com.br",
		PictureURL:   "example.com.br/profile-we.png",
		ProfileLink:  "example.com.br/FoowBar",
	}

	gotMovie.SetUpdateUser(&newUser)

	if gotMovie.UpdateUser != newUser {
		t.Errorf("\nWant: %v\nGot: %v\n\n", newUser, gotMovie.UpdateUser)
	}
}

func TestSetUpdateTime(t *testing.T) {
	gotMovie := NewValidMovie()

	oldTime := gotMovie.UpdateTime

	gotMovie.SetUpdateTime()

	oldTimeSeconds := time.Since(oldTime).Seconds()
	updatedTimeSeconds := time.Since(gotMovie.UpdateTime).Seconds()

	if oldTimeSeconds < updatedTimeSeconds {
		t.Errorf("Previous time '%v' should be lower than '%v'", oldTimeSeconds, updatedTimeSeconds)
	}
}

func TestValidMovie(t *testing.T) {
	gotMovie := NewValidMovie()

	gotMovie, _ = gotMovie.SetReleased("1996-12-19T16:39:57-08:00")

	gotMovie.
		SetTitle("Movie Title").
		SetRated("R").
		SetRunTime(19).
		SetDirector("Movie Director").
		SetWriter("Movie Writer")

	if gotMovie.IsValid() != nil {
		t.Errorf("\nWant: %v\nGot: %v\n\n", nil, gotMovie.IsValid())
	}
}

func TestInvalidMovieTitle(t *testing.T) {
	gotMovie := NewValidMovie()

	wantErr := errs.E(errs.Validation, errs.Parameter("title"), errs.MissingField("title"))

	if err := gotMovie.IsValid(); err.Error() != wantErr.Error() {
		t.Errorf("\nWant: %v\nGot: %v\n\n", wantErr.Error(), err.Error())
	}
}

func TestInvalidMovieRated(t *testing.T) {
	gotMovie := NewValidMovie()

	gotMovie.SetTitle("Movie Title")

	wantErr := errs.E(errs.Validation, errs.Parameter("rated"), errs.MissingField("Rated"))

	if err := gotMovie.IsValid(); err.Error() != wantErr.Error() {
		t.Errorf("\nWant: %v\nGot: %v\n\n", wantErr.Error(), err.Error())
	}
}

func TestInvalidMovieReleased(t *testing.T) {
	gotMovie := NewValidMovie()

	gotMovie.
		SetTitle("Movie Title").
		SetRated("R")

	wantErr := errs.E(errs.Validation, errs.Parameter("release_date"), "Released must have a value")

	if err := gotMovie.IsValid(); err.Error() != wantErr.Error() {
		t.Errorf("\nWant: %v\nGot: %v\n\n", wantErr.Error(), err.Error())
	}
}

func TestInvalidMovieRunTime(t *testing.T) {
	gotMovie := NewValidMovie()

	gotMovie, _ = gotMovie.SetReleased("1996-12-19T16:39:57-08:00")
	gotMovie.
		SetTitle("Movie Title").
		SetRated("R")

	wantErr := errs.E(errs.Validation, errs.Parameter("run_time"), "Run time must be greater than zero")

	if err := gotMovie.IsValid(); err.Error() != wantErr.Error() {
		t.Errorf("\nWant: %v\nGot: %v\n\n", wantErr.Error(), err.Error())
	}
}

func TestInvalidMovieDirector(t *testing.T) {
	gotMovie := NewValidMovie()

	gotMovie, _ = gotMovie.SetReleased("1996-12-19T16:39:57-08:00")
	gotMovie.
		SetTitle("Movie Title").
		SetRated("R").
		SetRunTime(19)

	wantErr := errs.E(errs.Validation, errs.Parameter("director"), errs.MissingField("Director"))

	if err := gotMovie.IsValid(); err.Error() != wantErr.Error() {
		t.Errorf("\nWant: %v\nGot: %v\n\n", wantErr.Error(), err.Error())
	}
}

func TestInvalidMovieWriter(t *testing.T) {
	gotMovie := NewValidMovie()

	gotMovie, _ = gotMovie.SetReleased("1996-12-19T16:39:57-08:00")
	gotMovie.
		SetTitle("Movie Title").
		SetRated("R").
		SetRunTime(19).
		SetDirector("Movie Director")

	wantErr := errs.E(errs.Validation, errs.Parameter("writer"), errs.MissingField("Writer"))

	if err := gotMovie.IsValid(); err.Error() != wantErr.Error() {
		t.Errorf("\nWant: %v\nGot: %v\n\n", wantErr.Error(), err.Error())
	}
}
