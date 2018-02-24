package middleware

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gilcrest/go-API-template/env"
	"github.com/rs/zerolog/log"
)

// LogResponse records and logs the response code, header and body details
// using an httptest.ResponseRecorder.  This function also manages the
// request/response timing
func LogResponse(env *env.Env, aud *APIAudit) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Print("Start LogResponse")
			defer log.Print("Finish LogResponse")

			startTimer(aud)

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)

			// copy everything from response recorder
			// to actual response writer
			for k, v := range rec.HeaderMap {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.Code)

			// pull out the response body and write it
			// back to the response writer
			b := rec.Body.Bytes()
			w.Write(b)

			stopTimer(aud)

			// write the data back to the recorder buffer as
			// it's needed for SetResponse
			rec.Body.Write(b)

			// set the response data in the APIAudit object
			err := setResponse(aud, rec)
			if err != nil {
				log.Error().Err(err).Msg("")
				http.Error(w, "Unable to set response", http.StatusBadRequest)
			}

			// call logRespDispatch to determine if and where to log
			err = logRespDispatch(env, aud)
			if err != nil {
				log.Error().Err(err).Msg("")
			}

		})
	}
}

// sets the start time in the APIAudit object
func startTimer(aud *APIAudit) {
	// set APIAudit TimeStarted to current time in UTC
	loc, _ := time.LoadLocation("UTC")
	aud.TimeStarted = time.Now().In(loc)
}

// stopTimer sets the stop time in the APIAudit object and
// subtracts the stop time from the start time to determine the
// service execution duration as this is after the response
// has been written and sent
func stopTimer(aud *APIAudit) {
	loc, _ := time.LoadLocation("UTC")
	aud.TimeFinished = time.Now().In(loc)
	duration := aud.TimeFinished.Sub(aud.TimeStarted)
	aud.Duration = duration
}

// SetResponse sets the response elements of the APIAudit payload
func setResponse(aud *APIAudit, rec *httptest.ResponseRecorder) error {
	// set ResponseCode from ResponseRecorder
	aud.ResponseCode = rec.Code

	// set Header JSON from Header map in ResponseRecorder
	headerJSON, err := convertHeader(rec.HeaderMap)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	aud.response.Header = headerJSON

	// Dump body to text using dumpBody function - need an http request
	// struct, so use httptest.NewRequest to get one
	req := httptest.NewRequest("POST", "http://example.com/foo", rec.Body)

	body, err := dumpBody(req)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	aud.response.Body = body

	return nil
}

// logRespDispatch determines which, if any, of the logging methods
// you wish to use will be employed
func logRespDispatch(env *env.Env, aud *APIAudit) error {
	if env.LogOpts.Log2StdOut.Response.Enable {
		logResp2Stdout(env, aud)
	}

	if env.LogOpts.Log2DB.Enable {
		err := logReqResp2Db(env, aud)
		if err != nil {
			log.Error().Err(err).Msg("")
			return err
		}
	}
	return nil
}

func logResp2Stdout(env *env.Env, aud *APIAudit) {
	logger := env.Logger

	logger.Debug().Msg("logResponse started")
	defer logger.Debug().Msg("logResponse ended")

	logger.Info().
		Str("request_id", aud.RequestID).
		Int("response_code", aud.ResponseCode).
		Str("response_header", aud.response.Header).
		Str("response_body", aud.response.Body).
		Msg("Response Sent")
}

// logReqResp2Db creates a record in the api.audit_log table
// using a stored function
func logReqResp2Db(env *env.Env, aud *APIAudit) error {

	var (
		rowsInserted int
		respHdr      interface{}
		respBody     interface{}
		reqHdr       interface{}
		reqBody      interface{}
	)

	// default reqHdr variable to nil
	// if the Request Header logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	reqHdr = nil
	if env.LogOpts.Log2DB.Request.Header {
		// This empty string to nil conversion is probably
		// not necessary, but just in case to avoid db exception
		reqHdr = strNil(aud.request.Header)
	}
	// default reqBody variable to nil
	// if the Request Body logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	reqBody = nil
	if env.LogOpts.Log2DB.Request.Body {
		reqBody = strNil(aud.request.Body)
	}
	// default respHdr variable to nil
	// if the Response Header logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	respHdr = nil
	if env.LogOpts.Log2DB.Response.Header {
		respHdr = strNil(aud.response.Header)
	}
	// default respBody variable to nil
	// if the Response Body logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	respBody = nil
	if env.LogOpts.Log2DB.Response.Body {
		respBody = strNil(aud.response.Body)
	}

	// Calls the BeginTx method of the LogDB opened database
	tx, err := env.DS.LogDb.BeginTx(aud.ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}

	// time.Duration is in nanoseconds,
	// need to do below math for milliseconds
	durMS := aud.Duration / time.Millisecond

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(aud.ctx, `select api.log_request (
		p_request_id => $1,
		p_request_timestamp => $2,
		p_response_code => $3,
		p_response_timestamp => $4,
		p_duration_in_millis => $5,
		p_protocol => $6,
		p_protocol_major => $7,
		p_protocol_minor => $8,
		p_request_method => $9,
		p_scheme => $10,
		p_host => $11,
		p_port => $12,
		p_path => $13,
		p_remote_address => $14,
		p_request_content_length => $15,
		p_request_header => $16,
		p_request_body => $17,
		p_response_header => $18,
		p_response_body => $19)`)

	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(aud.ctx,
		aud.RequestID,             //$1
		aud.TimeStarted,           //$2
		aud.ResponseCode,          //$3
		aud.TimeFinished,          //$4
		durMS,                     //$5
		aud.request.Proto,         //$6
		aud.request.ProtoMajor,    //$7
		aud.request.ProtoMinor,    //$8
		aud.request.Method,        //$9
		aud.request.Scheme,        //$10
		aud.request.Host,          //$11
		aud.request.Port,          //$12
		aud.request.Path,          //$13
		aud.request.RemoteAddr,    //$14
		aud.request.ContentLength, //$15
		reqHdr,   //$16
		reqBody,  //$17
		respHdr,  //$18
		respBody) //$19

	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&rowsInserted); err != nil {
			log.Error().Err(err).Msg("")
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}

	// If we have successfully written rows to the db
	// we commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}

	return nil

}

// strNil checks if the header field is an empty string
// (the empty value for the string type) and switches it to
// a nil.  An empty string is not allowed to be passed to a
// JSONB type in postgres, however, a nil works
func strNil(s string) interface{} {
	var v interface{}

	v = s
	if s == "" {
		v = nil
	}

	return v
}
