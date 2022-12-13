package diygoapi

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
)

func TestMovie_IsValid(t *testing.T) {
	c := qt.New(t)

	rd, _ := time.Parse(time.RFC3339, "1985-08-16T00:00:00Z")

	movieFunc := func() *Movie {
		return &Movie{
			ID:         uuid.New(),
			ExternalID: secure.NewID(),
			Title:      "The Return of the Living Dead",
			Rated:      "R",
			Released:   rd,
			RunTime:    91,
			Director:   "Dan O'Bannon",
			Writer:     "Russell Streiner",
		}
	}

	m1 := movieFunc()
	m2 := movieFunc()
	m2.ExternalID = nil
	m2a := movieFunc()
	m2a.ExternalID = secure.Identifier{}
	m3 := movieFunc()
	m3.Title = ""
	m4 := movieFunc()
	m4.Rated = ""
	m5 := movieFunc()
	m5.Released = time.Time{}
	m6 := movieFunc()
	m6.RunTime = 0
	m7 := movieFunc()
	m7.Director = ""
	m8 := movieFunc()
	m8.Writer = ""

	tests := []struct {
		name    string
		m       *Movie
		wantErr error
	}{
		{"typical no error", m1, nil},
		{"nil ExternalID", m2, errs.E(errs.Validation, errs.Parameter("extlID"), errs.MissingField("extlID"))},
		{"empty ExternalID", m2a, errs.E(errs.Validation, errs.Parameter("extlID"), errs.MissingField("extlID"))},
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
