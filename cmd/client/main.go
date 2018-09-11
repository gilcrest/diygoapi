package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gilcrest/httplog"
	"github.com/rs/zerolog/log"
)

// UserClient struct type holds the information about
// the REST API we are going to consume
type userClient struct {
	BaseURL    *url.URL
	UserAgent  string
	HTTPClient *http.Client
}

// request is the expected service request fields
type request struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	MobileID  string `json:"mobile_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// response is the expected service response fields
type response struct {
	Username       string         `json:"username"`
	MobileID       string         `json:"mobile_id"`
	Email          string         `json:"email"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	UpdateUserID   string         `json:"update_user_id"`
	UpdateUnixTime int64          `json:"created"`
	Audit          *httplog.Audit `json:"audit"`
}

// Create method sets up the request, then calls the do method of said request
// and returns the response body
func (c *userClient) Create(ctx context.Context, body *request) (*response, error) {

	// get a new http.Request struct from newRequest function
	req, err := c.newRequest("POST", "/api/v1/appuser", body)
	if err != nil {
		return nil, err
	}

	respBody := new(response)

	_, err = c.do(ctx, req, &respBody)

	return respBody, err
}

// newRequest generates an http.Request struct
// which will be used in the subsequent httpClient.Do method
func (c *userClient) newRequest(method, path string, body interface{}) (*http.Request, error) {

	// relative Path
	rel := &url.URL{Path: path}

	// ResolveReference takes the relative path along with the base URL and
	// returns the full URL
	u := c.BaseURL.ResolveReference(rel)

	// Declare an io.ReadWriter buffer
	var buf io.ReadWriter

	// Assume if the body parameter is not null that it's JSON
	//  and encode/stream it into the buffer initialized below
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	// NewRequest returns a pointer to an http.request struct
	//  to be used with the client.do method
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	// If body interface is not null, for now, assume JSON and
	//  set the proper http header as such
	if body != nil {
		// Tell the server that the data in the body of the request
		// is JSON
		req.Header.Set("Content-Type", "application/json")
	}

	// Tell the server that the client will accept JSON as the
	// response body
	req.Header.Set("Accept", "application/json")

	// Set the User-Agent
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

func (c *userClient) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {

	// Take the http.Request and change it's context to the context
	// passed into the parameter
	req = req.WithContext(ctx)

	// Send the http request to the server and receive a response
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()

	// json.NewDecoder returns a pointer to the Decoder type
	err = json.NewDecoder(resp.Body).Decode(v)

	return resp, err
}

func main() {

	// Initialize an empty context (context.Background()) and then add a
	// timeout to it
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// use the Parse method of the URL struct to return a properly formed
	//  URL struct
	u, err := url.Parse("http://127.0.0.1:8080/")
	if err != nil {
		log.Fatal().Err(err)
	}

	// Initialize the UserClient struct with the formed URL from above and a pointer to
	//  then default http client in the http package
	client := userClient{BaseURL: u, UserAgent: "gilcrest", HTTPClient: http.DefaultClient}

	usr := new(request)
	usr.Username = "repoMan"
	usr.Password = "wearYourSeatbelt"
	usr.MobileID = "(617) 302-7777"
	usr.Email = "repoman@alwaysintense.com"
	usr.FirstName = "Otto"
	usr.LastName = "Maddox"

	// clt.Create does the actual http POST to the endpoint to create an application user
	user, err := client.Create(ctx, usr)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Fatal().
				Err(err).
				Str("Timeout", "Response timed out")
		}
		log.Fatal().Err(err)
	}
	log.Print(user)
}
