package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/gilcrest/go-API-template/pkg/api/server/handlers"
	"github.com/gilcrest/go-API-template/pkg/env"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type APIAudit struct {
	ctx           context.Context
	RequestID     xid.ID    `json:"request_id"`
	TimeStarted   time.Time `json:"time_started"`
	TimeFinished  time.Time `json:"time_finished"`
	TimeInMillis  int       `json:"time_in_millis"`
	Proto         string    `json:"protocol"`
	ProtoMajor    int       `json:"protocol_major"`
	ProtoMinor    int       `json:"protocol_minor"`
	Method        string    `json:"request_method"`
	Scheme        string    `json:"scheme"`
	Host          string    `json:"host"`
	Port          string    `json:"port"`
	Path          string    `json:"path"`
	Header        string    `json:"header"`
	ContentLength int64     `json:"content_length"`
	RemoteAddr    string    `json:"remote_address"`
}

// LogRequest wraps several optional logging functions
//   printRequest - sends request output from httputil.DumpRequest to STDERR
//   logRequest - uses logger util to log requests
//   log2DB - logs request to relational database TODO - not yet implemented
func LogRequest(env *env.Env, h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			err := logSwitch(env, r)
			if err != nil {
				http.Error(w, "Unable to log request", http.StatusBadRequest)
				return
			}
			h.ServeHTTP(w, r) // call original
		})
}

func newAPIAudit(req *http.Request) (*APIAudit, error) {
	var scheme string
	host, port, err := net.SplitHostPort(req.Host)

	isHTTPS := req.TLS != nil

	if isHTTPS {
		scheme = "https"
	} else {
		scheme = "http"
	}

	jsonBytes, err := json.Marshal(req.Header)
	if err != nil {
		return nil, err
	}
	headerJSON := string(jsonBytes)

	apiAud := APIAudit{ctx: req.Context(), // retrieve the context from the http.Request
		RequestID:     xid.New(),
		TimeStarted:   time.Now(),
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		Method:        req.Method,
		Scheme:        scheme,
		Host:          host,
		Port:          port,
		Path:          req.URL.Path,
		Header:        headerJSON,
		ContentLength: req.ContentLength,
		RemoteAddr:    req.RemoteAddr}

	return &apiAud, nil

}

// logSwitch determines which, if any, of the logging methods
// you wish to use will be employed.  Using a cache mechanism I haven't
// implemented yet, you will be able to turn on/off these methods on demand
// as of right now, it's doing all of them, which is ridiculous
func logSwitch(env *env.Env, req *http.Request) error {
	// Check cached key:value pair to determine if printing/logging is on
	// for the service
	// TODO - Implement cache - maybe via Redis?
	var err error

	apiAudit, err := newAPIAudit(req)
	if err != nil {
		return err
	}

	// err = printRequest(req)
	// if err != nil {
	// 	return err
	// }
	// err = logRequest(env, req)
	// if err != nil {
	// 	return err
	// }

	err = logRequest2Db(env, apiAudit)
	if err != nil {
		return err
	}

	return nil

}

// PrintRequest wraps the call to httputil.DumpRequest
func printRequest(req *http.Request) error {

	// func DumpRequest(req *http.Request, body bool) ([]byte, error)
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return handlers.HTTPStatusError{Code: http.StatusBadRequest, Err: err}
	}
	fmt.Println(string(requestDump))
	return nil
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

func dumpBody(req *http.Request) ([]byte, error) {
	var err error
	save := req.Body
	save, req.Body, err = drainBody(req.Body)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return b.Bytes(), nil
}

func logBody(lgr *zap.Logger, req *http.Request) (*zap.Logger, error) {
	// func dumpBody(req *http.Request) ([]byte, error)
	requestDump, err := dumpBody(req)
	if err != nil {
		return nil, handlers.HTTPStatusError{Code: http.StatusBadRequest, Err: err}
	}
	lgr = lgr.With(zap.String("Body", string(requestDump)))
	return lgr, nil
}

func logHeader(lgr *zap.Logger, req *http.Request) (*zap.Logger, error) {

	for key, valSlice := range req.Header {
		for _, val := range valSlice {
			var i int
			i++
			header := fmt.Sprintf("%s: %s", key, val)
			lgr = lgr.With(zap.String(fmt.Sprintf("Header(%d)", i), header))
		}
	}
	return lgr, nil
}

func logTrailer(lgr *zap.Logger, req *http.Request) (*zap.Logger, error) {

	for key, valSlice := range req.Trailer {
		for _, val := range valSlice {
			var i int
			i++
			header := fmt.Sprintf("%s: %s", key, val)
			lgr = lgr.With(zap.String(fmt.Sprintf("Trailer(%d)", i), header))
		}
	}
	return lgr, nil
}

func logFormValues(lgr *zap.Logger, req *http.Request) (*zap.Logger, error) {

	var i int

	err := req.ParseForm()
	if err != nil {
		return nil, err
	}

	for key, valSlice := range req.Form {
		for _, val := range valSlice {
			i++
			formValue := fmt.Sprintf("%s: %s", key, val)
			lgr = lgr.With(zap.String(fmt.Sprintf("Form(%d)", i), formValue))
		}
	}

	for key, valSlice := range req.PostForm {
		for _, val := range valSlice {
			i++
			formValue := fmt.Sprintf("%s: %s", key, val)
			lgr = lgr.With(zap.String(fmt.Sprintf("PostForm(%d)", i), formValue))
		}
	}

	return lgr, nil
}

func logHostPort(lgr *zap.Logger, req *http.Request) (*zap.Logger, error) {
	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		return nil, handlers.HTTPStatusError{Code: http.StatusBadRequest, Err: err}
	}
	lgr = lgr.With(zap.String("Host", host))
	lgr = lgr.With(zap.String("Port", port))
	return lgr, nil
}

func logRequest(env *env.Env, req *http.Request) error {

	var err error

	logger := env.Logger
	defer env.Logger.Sync()

	logger.Debug("logRequest started")
	defer logger.Debug("logRequest ended")

	logger, err = logHostPort(logger, req)
	if err != nil {
		return err
	}
	logger, err = logHeader(logger, req)
	if err != nil {
		return err
	}
	logger, err = logBody(logger, req)
	if err != nil {
		return err
	}
	logger, err = logFormValues(logger, req)
	if err != nil {
		return err
	}
	logger, err = logTrailer(logger, req)
	if err != nil {
		return err
	}
	// TODO - determine what to log for TLS (*tls.ConnectionState)

	logger.Info("Request received",
		zap.String("HTTP method", req.Method),
		zap.String("URL Path", req.URL.Path[1:]),
		zap.String("URL", req.URL.String()),
		zap.String("Protocol", req.Proto),
		zap.Int("ProtoMajor", req.ProtoMajor),
		zap.Int("ProtoMinor", req.ProtoMinor),
		zap.Int64("Content Length", req.ContentLength),
		zap.String("Transfer-Encoding", strings.Join(req.TransferEncoding, ",")),
		zap.Bool("Close", req.Close),
		zap.String("RemoteAddr", req.RemoteAddr),
		zap.String("RequestURI", req.RequestURI),
	)

	return nil
}

// Creates a record in the appUser table using a stored function
func logRequest2Db(env *env.Env, req *APIAudit) error {

	var rowsInserted int

	// Calls the BeginTx method of the MainDb opened database
	tx, err := env.DS.MainDb.BeginTx(req.ctx, nil)
	if err != nil {
		return err
	}

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(req.ctx, `select api.log_request
		(
		p_request_id => $1,
		p_protocol => $2,
		p_protocol_major => $3,
		p_protocol_minor => $4,
		p_request_method => $5,
		p_scheme => $6,
		p_host => $7,
		p_port => $8,
		p_path => $9,
		p_header => $10,
		p_content_length => $11,
		p_remote_address => $12)`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(req.ctx,
		req.RequestID,
		req.Proto,
		req.ProtoMajor,
		req.ProtoMinor,
		req.Method,
		req.Scheme,
		req.Host,
		req.Port,
		req.Path,
		req.Header,
		req.ContentLength,
		req.RemoteAddr)

	if err != nil {
		return err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&rowsInserted); err != nil {
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
