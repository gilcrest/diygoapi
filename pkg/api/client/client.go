// Package client package sets up clients for calling the APIs in this app
// approach largely taken from Marcus Olsson
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gilcrest/go-API-template/pkg/domain/appUser"
)

// UserClient struct type holds the information about
// the REST API we are going to consume
type UserClient struct {
	BaseURL   *url.URL
	UserAgent string

	HTTPClient *http.Client
}

// Create method sets up the request, then calls the do method of said request
//  and returns the appUser.User returned in the response body
func (c *UserClient) Create(ctx context.Context, body *appUser.User) (*appUser.User, error) {

	// get a new http.Request struct from newRequest function
	req, err := c.newRequest("POST", "/api/v1/appUser", body)
	if err != nil {
		return nil, err
	}

	var respBody *appUser.User

	_, err = c.do(ctx, req, &respBody)

	return respBody, err
}

// newRequest generates an http.Request struct
// which will be used in the subsequent httpClient.Do method
func (c *UserClient) newRequest(method, path string, body interface{}) (*http.Request, error) {

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

func (c *UserClient) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {

	// Take the http.Request and change it's context to the context
	//  passed int the parameter
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
