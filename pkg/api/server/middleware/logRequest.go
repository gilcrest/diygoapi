package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gilcrest/go-API-template/pkg/env"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

// const (
// 	dumpReq   = "dumpRequest"
// 	logReq    = "logRequest"
// 	logResp   = "logResponse"
// 	logReq2db = "logReq2db"
// 	// m["reqwbody2db"] = true
// 	// m["resp2db"] = true
// 	// m["respwbody2db"] = true
// 	// m["limited2db"] = true
// )

// APIAudit struct holds the http request and other attributes needed
// for auditing an http request
type APIAudit struct {
	ctx          context.Context
	RequestID    string        `json:"request_id"`
	TimeStarted  time.Time     `json:"time_started"`
	TimeFinished time.Time     `json:"time_finished"`
	TimeInMillis time.Duration `json:"time_in_millis"`
	ResponseCode int           `json:"response_code"`
	request
	response request
}

type request struct {
	Proto            string `json:"protocol"`
	ProtoMajor       int    `json:"protocol_major"`
	ProtoMinor       int    `json:"protocol_minor"`
	Method           string `json:"request_method"`
	Scheme           string `json:"scheme"`
	Host             string `json:"host"`
	Port             string `json:"port"`
	Path             string `json:"path"`
	Header           string `json:"header"`
	Body             string `json:"body"`
	ContentLength    int64  `json:"content_length"`
	TransferEncoding string `json:"transfer_encoding"`
	Close            bool   `json:"close"`
	Trailer          string `json:"trailer"`
	RemoteAddr       string `json:"remote_address"`
	RequestURI       string `json:"request_uri"`
}

// LogRequest wraps several optional logging functions
//   printRequest - sends request output from httputil.DumpRequest to STDERR
//   logRequest - uses logger util to log requests
//   log2DB - logs request to relational database TODO - not yet implemented
func LogRequest(env *env.Env, aud *APIAudit) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Print("Start LogRequest")
			defer log.Print("Finish LogRequest")
			err := logReqDispatch(env, aud, r)
			if err != nil {
				http.Error(w, "Unable to log request", http.StatusBadRequest)
				return
			}
			h.ServeHTTP(w, r) // call original
		})
	}
}

// logReqDispatch determines which, if any, of the logging methods
// you wish to use will be employed.  Using a cache mechanism I haven't
// implemented yet, you will be able to turn on/off these methods on demand
// as of right now, it's doing all of them, which is ridiculous
func logReqDispatch(env *env.Env, aud *APIAudit, req *http.Request) error {

	var err error

	err = SetRequest(aud, req)
	if err != nil {
		return err
	}

	if env.LogOpts.DumpRequest.Write {
		// func DumpRequest(req *http.Request, body bool) ([]byte, error)
		requestDump, err := httputil.DumpRequest(req, env.LogOpts.DumpRequest.Body)
		if err != nil {
			return err
		}
		log.Print(string(requestDump))
		return nil
	}

	if env.LogOpts.Log2StdOut.Request.Write {
		err = logRequest(env, aud)
		if err != nil {
			return err
		}
	}

	// if logSwitch(env, logReq2db) {
	// 	err = logRequest2Db(env, aud)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil

}

// func logSwitch(env *env.Env, logTest string) bool {

// 	// If nothing is in the LogMap, then return false
// 	if len(env.LogMap) == 0 {
// 		return false
// 	}

// 	// for dumprequest, validate if true in map
// 	if logTest == dumpReq {
// 		if mv := env.LogMap[dumpReq]; mv {
// 			return true
// 		}
// 	}

// 	// for logRequest, validate if true in map
// 	if logTest == logReq {
// 		if mv := env.LogMap[logReq]; mv {
// 			return true
// 		}
// 	}

// 	// for logResponse, validate if true in map
// 	if logTest == logResp {
// 		if mv := env.LogMap[logResp]; mv {
// 			return true
// 		}
// 	}

// 	return false
// }

// SetRequest populates several core fields (TimeStarted, ctx and RequestID)
// for the APIAudit struct being passed into the function
func SetRequest(aud *APIAudit, req *http.Request) error {

	var (
		scheme string
	)

	// split host and port out for cleaner logging
	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		return err
	}

	// determine if the request is an HTTPS request
	isHTTPS := req.TLS != nil

	if isHTTPS {
		scheme = "https"
	} else {
		scheme = "http"
	}

	// convert the Header map from the request to a JSON string
	headerJSON, err := convertHeader(req.Header)
	if err != nil {
		return err
	}

	// convert the Trailer map from the request to a JSON string
	trailerJSON, err := convertHeader(req.Trailer)
	if err != nil {
		return err
	}

	// get byte Array representation of guid from xid package (12 bytes)
	guid := xid.New()

	// use the String method of the guid object to convert byte array to string (20 bytes)
	rID := guid.String()

	body, err := dumpBody(req)
	if err != nil {
		return err
	}

	aud.ctx = req.Context() // retrieve the context from the http.Request
	aud.RequestID = rID
	aud.request.Proto = req.Proto
	aud.request.ProtoMajor = req.ProtoMajor
	aud.request.ProtoMinor = req.ProtoMinor
	aud.request.Method = req.Method
	aud.request.Scheme = scheme
	aud.request.Host = host
	aud.request.Body = body
	aud.request.Port = port
	aud.request.Path = req.URL.Path
	aud.request.Header = headerJSON
	aud.request.ContentLength = req.ContentLength
	aud.request.TransferEncoding = strings.Join(req.TransferEncoding, ",")
	aud.request.Close = req.Close
	aud.request.Trailer = trailerJSON
	aud.request.RemoteAddr = req.RemoteAddr
	aud.request.RequestURI = req.RequestURI

	return nil

}

// SetResponse sets the response elements of the APIAudit payload
func SetResponse(aud *APIAudit, rec *httptest.ResponseRecorder) error {
	// set ResponseCode from ResponseRecorder
	aud.ResponseCode = rec.Code

	// set Header JSON from Header map in ResponseRecorder
	headerJSON, err := convertHeader(rec.HeaderMap)
	if err != nil {
		return err
	}
	aud.response.Header = headerJSON

	// Dump body to text using dumpBody function - need an http request
	// struct, so use httptest.NewRequest to get one
	req := httptest.NewRequest("POST", "http://example.com/foo", rec.Body)

	body, err := dumpBody(req)
	if err != nil {
		return err
	}
	aud.response.Body = body

	return nil
}

// convertHeader returns a JSON string representation of an http.Header map
func convertHeader(hdr http.Header) (string, error) {
	// convert the http.Header map from the request to a slice of bytes
	jsonBytes, err := json.Marshal(hdr)
	if err != nil {
		return "", err
	}

	// convert the slice of bytes to a JSON string
	headerJSON := string(jsonBytes)

	return headerJSON, nil

}

// drainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes.
//
// It returns an error if the initial slurp of all bytes fails. It does not attempt
// to make the returned ReadClosers have identical error-matching behavior.
// Function lifted straight from net/http/httputil package
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func dumpBody(req *http.Request) (string, error) {
	var err error
	save := req.Body
	save, req.Body, err = drainBody(req.Body)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer

	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"

	if req.Body != nil {
		var dest io.Writer = &b
		if chunked {
			dest = httputil.NewChunkedWriter(dest)
		}
		_, err = io.Copy(dest, req.Body)
		if chunked {
			dest.(io.Closer).Close()
			io.WriteString(&b, "\r\n")
		}
	}

	req.Body = save
	if err != nil {
		return "", err
	}
	return string(b.Bytes()), nil
}

// func logFormValues(lgr zerolog.Logger, req *http.Request) (zerolog.Logger, error) {

// 	var i int

// 	err := req.ParseForm()
// 	if err != nil {
// 		return nil, err
// 	}

// 	for key, valSlice := range req.Form {
// 		for _, val := range valSlice {
// 			i++
// 			formValue := fmt.Sprintf("%s: %s", key, val)
// 			lgr = lgr.With().Str(fmt.Sprintf("Form(%d)", i), formValue))
// 		}
// 	}

// 	for key, valSlice := range req.PostForm {
// 		for _, val := range valSlice {
// 			i++
// 			formValue := fmt.Sprintf("%s: %s", key, val)
// 			lgr = lgr.With(Str(fmt.Sprintf("PostForm(%d)", i), formValue))
// 		}
// 	}

// 	return lgr, nil
// }

func logRequest(env *env.Env, aud *APIAudit) error {

	//var err error

	log := env.Logger

	log.Debug().Msg("logRequest started")
	defer log.Debug().Msg("logRequest ended")

	// logger, err = logFormValues(logger, req)
	// if err != nil {
	// 	return err
	// }

	// All header key:value pairs written to JSON
	if env.LogOpts.Log2StdOut.Request.Header {
		log = log.With().Str("header_json", aud.request.Header).Logger()
	}

	if env.LogOpts.Log2StdOut.Request.Body {
		log = log.With().Str("body", aud.request.Body).Logger()
	}

	log.Info().
		Str("request_id", aud.RequestID).
		Str("method", aud.request.Method).
		// most url.URL components split out
		Str("scheme", aud.request.Scheme).
		Str("host", aud.request.Host).
		Str("port", aud.request.Port).
		Str("path", aud.request.Path).
		// The protocol version for incoming server requests.
		Str("protocol", aud.request.Proto).
		Int("proto_major", aud.request.ProtoMajor).
		Int("proto_minor", aud.request.ProtoMinor).
		Int64("Content Length", aud.request.ContentLength).
		Str("Transfer-Encoding", aud.request.TransferEncoding).
		Bool("Close", aud.request.Close).
		Str("RemoteAddr", aud.request.RemoteAddr).
		Str("RequestURI", aud.request.RequestURI).
		Msg("Request received")

	return nil
}

// Creates a record in the appUser table using a stored function
func logRequest2Db(env *env.Env, aud *APIAudit) error {

	var (
		rowsInserted int
		respHdr      interface{}
		reqHdr       interface{}
	)

	respHdr = headerNil(aud.response.Header)
	reqHdr = headerNil(aud.request.Header)

	// Calls the BeginTx method of the MainDb opened database
	tx, err := env.DS.MainDb.BeginTx(aud.ctx, nil)
	if err != nil {
		return err
	}

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(aud.ctx, `select api.log_request
		(
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
		p_request_header => $15,
		p_request_content_length => $16,
		p_request_body => $17,
		p_response_header => $18,
		p_response_body => $19)`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(aud.ctx,
		aud.RequestID,          //$1
		aud.TimeStarted,        //$2
		aud.ResponseCode,       //$3
		aud.TimeFinished,       //$4
		aud.TimeInMillis,       //$5
		aud.request.Proto,      //$6
		aud.request.ProtoMajor, //$7
		aud.request.ProtoMinor, //$8
		aud.request.Method,     //$9
		aud.request.Scheme,     //$10
		aud.request.Host,       //$11
		aud.request.Port,       //$12
		aud.request.Path,       //$13
		aud.request.RemoteAddr, //$14
		reqHdr,                 //$15
		aud.request.ContentLength, //$16
		aud.request.Body,          //$17
		respHdr,                   //$18
		aud.response.Body)         //$19

	if err != nil {
		log.Print(err)
		return err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&rowsInserted); err != nil {
			log.Print(err)
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	// If we have successfully written rows to the db
	// we commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

// headerNil checks if the header field is an empty string
// (the empty value for the string type) and switches it to
// a nil.  An empty string is not allowed to be passed to a
// JSONB type in postgres, however, a nil works
func headerNil(hdr string) interface{} {
	var respHdr interface{}

	if hdr == "" {
		respHdr = nil
	} else {
		respHdr = hdr
	}

	return respHdr
}

func logResponse(env *env.Env, aud *APIAudit) error {

	//var err error

	logger := env.Logger

	logger.Debug().Msg("logResponse started")
	defer logger.Debug().Msg("logResponse ended")

	// logger, err = logFormValues(logger, req)
	// if err != nil {
	// 	return err
	// }

	logger.Info().
		Str("request_id", aud.RequestID).
		Int("response_code", aud.ResponseCode).
		Str("header_json", aud.response.Header).
		Str("body", aud.response.Body).
		Msg("Response Sent")

	return nil
}
