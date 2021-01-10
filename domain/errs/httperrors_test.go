package errs

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/logger"
)

func Test_httpErrorStatusCode(t *testing.T) {
	type args struct {
		k Kind
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"Unauthenticated", args{k: Unauthenticated}, http.StatusUnauthorized},
		{"Unauthorized", args{k: Unauthorized}, http.StatusForbidden},
		{"Permission", args{k: Permission}, http.StatusForbidden},
		{"Exist", args{k: Exist}, http.StatusBadRequest},
		{"Invalid", args{k: Invalid}, http.StatusBadRequest},
		{"NotExist", args{k: NotExist}, http.StatusBadRequest},
		{"Private", args{k: Private}, http.StatusBadRequest},
		{"BrokenLink", args{k: BrokenLink}, http.StatusBadRequest},
		{"Validation", args{k: Validation}, http.StatusBadRequest},
		{"InvalidRequest", args{k: InvalidRequest}, http.StatusBadRequest},
		{"InvalidRequest", args{k: InvalidRequest}, http.StatusBadRequest},
		{"InvalidRequest", args{k: InvalidRequest}, http.StatusBadRequest},
		{"InvalidRequest", args{k: InvalidRequest}, http.StatusBadRequest},
		{"InvalidRequest", args{k: InvalidRequest}, http.StatusBadRequest},
		{"Other", args{k: Other}, http.StatusInternalServerError},
		{"IO", args{k: IO}, http.StatusInternalServerError},
		{"Internal", args{k: Internal}, http.StatusInternalServerError},
		{"Database", args{k: Database}, http.StatusInternalServerError},
		{"Unanticipated", args{k: Unanticipated}, http.StatusInternalServerError},
		{"Default", args{k: 99}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := httpErrorStatusCode(tt.args.k); got != tt.want {
				t.Errorf("httpErrorStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPErrorResponse_StatusCode(t *testing.T) {

	type args struct {
		w   *httptest.ResponseRecorder
		l   zerolog.Logger
		err error
	}

	var b bytes.Buffer
	l := logger.NewLogger(&b, false)

	tests := []struct {
		name string
		args args
		want int
	}{
		{"empty", args{httptest.NewRecorder(), l, &Error{}}, http.StatusInternalServerError},
		{"unauthenticated", args{httptest.NewRecorder(), l, E(Unauthenticated)}, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HTTPErrorResponse(tt.args.w, l, tt.args.err)
			if got := tt.args.w.Result().StatusCode; got != tt.want {
				t.Errorf("httpErrorStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
