package handler

import (
	"encoding/json"
	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppHandler_AddMovie(t *testing.T) {
	type requestFields struct {
		HTTPMethod  string
		URL         string
		RequestBody io.Reader
	}
	type responseFields struct {
		Status       int
		ResponseBody string
	}

	nilBodyRequest := requestFields{
		HTTPMethod:  "POST",
		URL:         "/api/v1/movies",
		RequestBody: nil,
	}

	er := errs.ErrResponse{Error: errs.ServiceError{
		Kind:    errs.InvalidRequest.String(),
		Message: "Request Body cannot be empty",
	}}

	errJSON, _ := json.Marshal(er)

	nilBodyResponse := responseFields{
		Status:       http.StatusBadRequest,
		ResponseBody: string(errJSON),
	}

	tests := []struct {
		name          string
		requestFields requestFields
		want          responseFields
	}{
		{"nil body", nilBodyRequest, nilBodyResponse},
		//{"mock post", }
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
			// pass 'nil' as the third parameter.
			req, err := http.NewRequest(tt.requestFields.HTTPMethod, tt.requestFields.URL, tt.requestFields.RequestBody)
			if err != nil {
				t.Fatal(err)
			}

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			rr := httptest.NewRecorder()

			a := app.NewApplication(app.Local, datastore.NewMockDatastore(), app.NewLogger(zerolog.DebugLevel))
			appHandler := NewAppHandler(a)
			handler := http.HandlerFunc(appHandler.AddMovie)

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			if status := rr.Code; status != tt.want.Status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Check the response body is what we expect.
			// strings.TrimSpace removes the end of line chars
			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.want.ResponseBody) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.want.ResponseBody)
			}

		})
	}
}
