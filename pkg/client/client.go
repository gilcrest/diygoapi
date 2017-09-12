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

	"github.com/gilcrest/go-API-template/pkg/appUser"
)

// UserClient is a ...
type UserClient struct {
	BaseURL   *url.URL
	UserAgent string

	HTTPClient *http.Client
}

// Create method is used to generate the
func (c *UserClient) Create(ctx context.Context, body appUser.User) (string, error) {

	// get a new http.Request struct from newRequest function
	req, err := c.newRequest("POST", "/api/v1/appUser/create", body)
	if err != nil {
		return "", err
	}

	var response string
	_, err = c.do(ctx, req, &response)

	return response, err
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
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

func (c *UserClient) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {

	req = req.WithContext(ctx)

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
	//err = json.NewDecoder(resp.Body).Decode(v)
	v = resp.Body

	return resp, err
}
