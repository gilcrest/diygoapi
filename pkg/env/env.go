// Package env has a type to store common environment related items
// sql db, logger, etc. as well as a constructor-like function to instantiate it
package env

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/gilcrest/go-API-template/pkg/datastore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Env type stores common environment related items
type Env struct {
	DS      *datastore.Datastore
	Logger  zerolog.Logger
	LogOpts *HTTPLogOpts
}

// HTTPLogOpts represent HTTP Logging Options
type HTTPLogOpts struct {
	HTTPUtil   *httputil   `json:"httputil"`
	Log2StdOut *log2StdOut `json:"log_json"`
	Log2DB     *log2DB     `json:"log_2DB"`
}

type httputil struct {
	DumpRequest *dumpRequest
}

type dumpRequest struct {
	Enable bool `json:"enable"`
	Body   bool `json:"body"`
}

type log2StdOut struct {
	Request  *l2SOpt
	Response *l2SOpt
}

// l2SOpt is the log2StdOut Options
// Enable should be true if you want to write the log, set
// the rOpt Header and Body accordingly if you want to write those
type l2SOpt struct {
	Enable  bool `json:"enable"`
	Options *rOpt
}

type log2DB struct {
	Enable   bool `json:"enable"`
	Request  *rOpt
	Response *rOpt
}

// rOpt is the http request/response logging options
// choose whether you want to log the http headers or body
// by setting either value to true
type rOpt struct {
	Header bool `json:"header"`
	Body   bool `json:"body"`
}

// NewEnv constructs Env type to be passed around to functions
func NewEnv() (*Env, error) {

	// setup logger
	logger := newLogger()
	// if err != nil {
	// 	return nil, err
	// }

	// open db connection pools
	ds, err := datastore.NewDatastore()
	if err != nil {
		return nil, err
	}

	// get logMap with initialized values
	lopts := newHTTPLogOpts()

	environment := &Env{Logger: logger, DS: ds, LogOpts: lopts}

	return environment, nil
}

func newHTTPLogOpts() *HTTPLogOpts {

	raw, err := ioutil.ReadFile("../go-API-template/pkg/fileInput/httpLogOpt.json")
	if err != nil {
		log.Fatal().Err(err)
	}

	var l HTTPLogOpts
	json.Unmarshal(raw, &l)

	return &l
}

func newLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = ""
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return lgr
}
