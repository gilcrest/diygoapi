package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gilcrest/go-API-template/pkg/appUser"
	"github.com/gilcrest/go-API-template/pkg/client"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	u, err := url.Parse("http://127.0.0.1:8080/")
	if err != nil {
		log.Fatal(err)
	}

	var dflt = http.DefaultClient
	clt := client.UserClient{BaseURL: u, UserAgent: "Gilcrest", HTTPClient: dflt}

	usr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}

	user, err := clt.Create(ctx, usr)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Fatal("Response timed out, do something different here if you want to...")
		}
		log.Fatal(err)
	}
	fmt.Print(user)
}
