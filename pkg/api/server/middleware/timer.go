package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Timer gets the time taken to process the request and form the response
// Timer is not the real time between writes, but is accurate enough for me (for now)
func Timer(aud *APIAudit) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Print("Start Timer")
			defer log.Print("Finish Timer")
			// set APIAudit TimeStarted to current time in UTC
			loc, _ := time.LoadLocation("UTC")
			aud.TimeStarted = time.Now().In(loc)
			log.Printf("aud.TimeStarted = %s\n", aud.TimeStarted)
			h.ServeHTTP(w, r)
			aud.TimeFinished = time.Now().In(loc)
			duration := aud.TimeFinished.Sub(aud.TimeStarted)
			aud.TimeInMillis = duration
			log.Printf("aud.TimeFinished = %s\n", aud.TimeFinished)
			log.Printf("aud.TimeInMillis = %s\n", aud.TimeInMillis)
		})
	}
}
