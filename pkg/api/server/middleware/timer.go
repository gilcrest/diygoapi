package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// Timer gets the time taken to process the request and form the response
// Timer is not the real time between writes, but is accurate enough for me (for now)
func Timer(aud *APIAudit) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Start Timer")
			defer fmt.Println("Finish Timer")
			// set APIAudit TimeStarted to current time in UTC
			loc, _ := time.LoadLocation("UTC")
			aud.TimeStarted = time.Now().In(loc)
			fmt.Printf("aud.TimeStarted = %s\n", aud.TimeStarted)
			h.ServeHTTP(w, r)
			aud.TimeFinished = time.Now().In(loc)
			duration := aud.TimeFinished.Sub(aud.TimeStarted)
			aud.TimeInMillis = duration
			fmt.Printf("aud.TimeFinished = %s\n", aud.TimeFinished)
			fmt.Printf("aud.TimeInMillis = %s\n", aud.TimeInMillis)
		})
	}
}
