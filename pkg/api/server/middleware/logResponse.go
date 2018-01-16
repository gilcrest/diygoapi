package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gilcrest/go-API-template/pkg/env"
)

// func NewDefaultHandler(h http.Handler, env *env.Env, aud *APIAudit) http.Handler {
// 	fmt.Println("Start NewDefaultHandler")
// 	defer fmt.Println("Finish NewDefaultHandler")
// 	return LogResponse(Timer(h, aud), env, aud)
// }

func LogResponse(h http.Handler, env *env.Env, aud *APIAudit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Start LogResponse")
		defer fmt.Println("Finish LogResponse")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		// copy everything from response recorder
		// to actual response writer
		for k, v := range rec.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		rec.Body.WriteTo(w)

		// var err error

		// err = SetResponse(aud, rec)
		// if err != nil {
		// 	fmt.Println("TODO FIX THIS ERROR")
		// }

		//fmt.Printf("%+v\n", aud)
		// err = logRespDispatch(env, aud)
		// if err != nil {
		// 	fmt.Println("TODO FIX THIS ERROR")
		// }

	})
}

// logRespDispatch determines which, if any, of the logging methods
// you wish to use will be employed.  Using a cache mechanism I haven't
// implemented yet, you will be able to turn on/off these methods on demand
// as of right now, it's doing all of them, which is ridiculous
func logRespDispatch(env *env.Env, aud *APIAudit) error {

	// Check cached key:value pair to determine if printing/logging is on
	// for the service
	// TODO - Implement cache - maybe via Redis?

	// var (
	// 	err error
	// )

	// if logSwitch(env, logResp) {
	// 	err = logResponse(env, aud)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil

}