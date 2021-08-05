package movie

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/domain/user/usertest"
)

func TestNewMovie(t *testing.T) {
	c := qt.New(t)

	id := uuid.New()
	extlID := "ExternalID"

	m := &Movie{
		ID:         id,
		ExternalID: extlID,
		Title:      "",
		Rated:      "",
		Released:   time.Time{},
		RunTime:    0,
		Director:   "",
		Writer:     "",
		CreateUser: usertest.NewUser(t),
		CreateTime: time.Now().UTC(),
		UpdateUser: usertest.NewUser(t),
		UpdateTime: time.Now().UTC(),
	}

	ignoreFields := cmpopts.IgnoreFields(Movie{}, "CreateTime", "UpdateTime")
	within1Second := cmpopts.EquateApproxTime(time.Second)

	type args struct {
		id     uuid.UUID
		extlID string
		u      user.User
	}
	tests := []struct {
		name    string
		args    args
		want    *Movie
		wantErr error
	}{
		{"typical", args{id, extlID, usertest.NewUser(t)}, m, nil},
		{"nil uuid", args{uuid.Nil, extlID, usertest.NewUser(t)}, nil, errs.E(errs.Validation, errs.Parameter("ID"), errs.MissingField("ID"))},
		{"empty External ID", args{id, "", usertest.NewUser(t)}, nil, errs.E(errs.Validation, errs.Parameter("extlID"), errs.MissingField("extlID"))},
		{"invalid User", args{id, extlID, usertest.NewInvalidUser(t)}, nil, errs.E(errs.Validation, errs.Parameter("User"), "User is invalid")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMovie(tt.args.id, tt.args.extlID, tt.args.u)
			if (err != nil) && (tt.wantErr == nil) {
				t.Errorf("NewMovie() error = %v, nil expected", err)
				return
			}
			c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), tt.wantErr)
			if got != nil {
				c.Assert(got, qt.CmpEquals(ignoreFields), tt.want)
				c.Assert(got.CreateTime, qt.CmpEquals(within1Second), tt.want.CreateTime)
				c.Assert(got.UpdateTime, qt.CmpEquals(within1Second), tt.want.UpdateTime)
			}
		})
	}
}

func TestSetExternalID(t *testing.T) {
	c := qt.New(t)

	got := new(Movie).SetExternalID("ExternalID")
	want := &Movie{ExternalID: "ExternalID"}

	c.Assert(got, qt.DeepEquals, want)
}

func TestSetTitle(t *testing.T) {
	c := qt.New(t)

	got := new(Movie).SetTitle("The Return of the Living Dead")
	want := &Movie{Title: "The Return of the Living Dead"}

	c.Assert(got, qt.DeepEquals, want)
}

func TestSetRated(t *testing.T) {
	c := qt.New(t)

	got := new(Movie).SetRated("R")
	want := &Movie{Rated: "R"}

	c.Assert(got, qt.DeepEquals, want)
}

func TestMovie_SetReleased(t *testing.T) {
	c := qt.New(t)

	type args struct {
		r string
	}

	tme, _ := time.Parse(time.RFC3339, "1985-08-16T00:00:00Z")

	og := new(Movie)
	wantMovie := &Movie{
		Released: tme,
	}

	tests := []struct {
		name    string
		ogMovie *Movie
		args    args
		want    *Movie
		wantErr bool
	}{
		{"typical", og, args{"1985-08-16T00:00:00Z"}, wantMovie, false},
		{"empty string", og, args{""}, wantMovie, true},
		{"the year 20000", og, args{"20000-08-16T00:00:00Z"}, wantMovie, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.ogMovie
			got, err := m.SetReleased(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetReleased() error = %v, nil expected", err)
				return
			}
			if got != nil {
				c.Assert(got, qt.DeepEquals, tt.want)
				return
			}
			c.Assert(err, qt.Not(qt.IsNil))
		})
	}
}

func TestSetRunTime(t *testing.T) {
	c := qt.New(t)

	got := new(Movie).SetRunTime(91)
	want := &Movie{RunTime: 91}

	c.Assert(got, qt.DeepEquals, want)
}

func TestSetDirector(t *testing.T) {
	c := qt.New(t)

	got := new(Movie).SetDirector(`Dan O'Bannon'`)
	want := &Movie{Director: `Dan O'Bannon'`}

	c.Assert(got, qt.DeepEquals, want)
}

func TestSetWriter(t *testing.T) {
	c := qt.New(t)

	got := new(Movie).SetWriter("Russell Streiner")
	want := &Movie{Writer: "Russell Streiner"}

	c.Assert(got, qt.DeepEquals, want)
}

func TestSetUpdateUser(t *testing.T) {
	c := qt.New(t)

	u := usertest.NewUser(t)

	got := new(Movie).SetUpdateUser(u)
	want := &Movie{UpdateUser: u}

	c.Assert(got, qt.DeepEquals, want)
}

func TestSetUpdateTime(t *testing.T) {
	// initialize quicktest checker
	c := qt.New(t)

	ogt := time.Now().UTC()

	// get a new movie
	m := new(Movie)

	// Call SetUpdateTime to update to now in utc
	m.SetUpdateTime()

	// should be within 1 second of ogt
	within1Second := cmpopts.EquateApproxTime(time.Second)

	c.Assert(ogt, qt.CmpEquals(within1Second), m.UpdateTime)
}

func TestMovie_IsValid(t *testing.T) {
	c := qt.New(t)

	rd, _ := time.Parse(time.RFC3339, "1985-08-16T00:00:00Z")

	movieFunc := func() *Movie {
		return &Movie{
			ID:         uuid.New(),
			ExternalID: "TROTLD",
			Title:      "The Return of the Living Dead",
			Rated:      "R",
			Released:   rd,
			RunTime:    91,
			Director:   "Dan O'Bannon",
			Writer:     "Russell Streiner",
			CreateUser: usertest.NewUser(t),
			CreateTime: time.Now().UTC(),
			UpdateUser: usertest.NewUser(t),
			UpdateTime: time.Now().UTC(),
		}
	}

	m1 := movieFunc()
	m2 := movieFunc().SetExternalID("")
	m3 := movieFunc().SetTitle("")
	m4 := movieFunc().SetRated("")
	m5, err := movieFunc().SetReleased("0001-01-01T00:00:00Z")
	c.Assert(err, qt.IsNil)
	m6 := movieFunc().SetRunTime(0)
	m7 := movieFunc().SetDirector("")
	m8 := movieFunc().SetWriter("")

	tests := []struct {
		name    string
		m       *Movie
		wantErr error
	}{
		{"typical no error", m1, nil},
		{"empty ExternalID", m2, errs.E(errs.Validation, errs.Parameter("extlID"), errs.MissingField("extlID"))},
		{"empty Title", m3, errs.E(errs.Validation, errs.Parameter("title"), errs.MissingField("title"))},
		{"empty Rated", m4, errs.E(errs.Validation, errs.Parameter("rated"), errs.MissingField("rated"))},
		{"zero Released", m5, errs.E(errs.Validation, errs.Parameter("release_date"), "release_date must have a value")},
		{"zero RunTime", m6, errs.E(errs.Validation, errs.Parameter("run_time"), "run_time must be greater than zero")},
		{"empty Director", m7, errs.E(errs.Validation, errs.Parameter("director"), errs.MissingField("director"))},
		{"empty Writer", m8, errs.E(errs.Validation, errs.Parameter("writer"), errs.MissingField("writer"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValidErr := tt.m.IsValid()
			if (isValidErr != nil) && (tt.wantErr == nil) {
				t.Errorf("IsValid() error = %v; nil expected", isValidErr)
				return
			}
			c.Assert(isValidErr, qt.CmpEquals(cmp.Comparer(errs.Match)), tt.wantErr)
		})
	}
}
