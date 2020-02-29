package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gilcrest/errs"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
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

	var emptyBody []byte
	nilBodyRequest := requestFields{
		HTTPMethod:  "POST",
		URL:         "/api/v1/movies",
		RequestBody: bytes.NewBuffer(emptyBody),
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

	requestBody := []byte(`{
		"title": "Repo Man",
		"year": 1984,
		"rated": "R",
		"release_date": "1984-03-02T00:00:00Z",
		"run_time": 92,
		"director": "Alex Cox",
		"writer": "Alex Cox"
	}`)

	reqf := requestFields{
		HTTPMethod:  "POST",
		URL:         "/api/v1/movies",
		RequestBody: bytes.NewBuffer(requestBody),
	}

	resf := responseFields{
		Status:       http.StatusOK,
		ResponseBody: `{"data":{"external_id":"mlPb1YimScrEsmJJa3Xd","title":"Repo Man","year":1984,"rated":"R","release_date":"1984-03-02T00:00:00Z","run_time":92,"director":"Alex Cox","writer":"Alex Cox","create_timestamp":"2020-02-25T00:00:00Z","update_timestamp":"2020-02-25T00:00:00Z"}}`,
	}

	tests := []struct {
		name          string
		requestFields requestFields
		want          responseFields
	}{
		{"empty body", nilBodyRequest, nilBodyResponse},
		{"post", reqf, resf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request
			req, err := http.NewRequest(tt.requestFields.HTTPMethod, tt.requestFields.URL, tt.requestFields.RequestBody)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
			rr := httptest.NewRecorder()

			a := app.NewMockedApplication(app.Local, datastore.NewMockDatastore(), app.NewLogger(zerolog.DebugLevel))
			appHandler := NewAppHandler(a)
			handler := http.HandlerFunc(appHandler.AddMovie)

			// Gorilla handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in the Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			if status := rr.Code; status != tt.want.Status {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.want.Status)
			}

			// Check the response body is what we expect.
			// strings.TrimSpace removes the end of line chars
			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.want.ResponseBody) {
				t.Errorf("handler returned unexpected body:\n got\n\t %v want\n\t %v",
					rr.Body.String(), tt.want.ResponseBody)
			}

		})
	}
}
