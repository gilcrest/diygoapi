package main

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/gilcrest/go-API-template/pkg/api/client"
	"github.com/gilcrest/go-API-template/pkg/domain/appUser"
	"github.com/rs/zerolog/log"
)

func main() {

	// Initialize an empty context (context.Background()) and then add a
	//  timeout to it
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
	clt := client.UserClient{BaseURL: u, UserAgent: "Gilcrest", HTTPClient: http.DefaultClient}

	// Initialize an instance of appUser.User
	usr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}

	// clt.Create does the actual http POST to the endpoint to create an application user
	user, err := clt.Create(ctx, &usr)
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
