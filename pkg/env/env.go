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
	DumpRequest dumpReqOpts `json:"dump_request"`
	Log2StdOut  reqResp     `json:"log_json"`
	Log2DB      reqResp     `json:"log_2DB"`
}

type reqResp struct {
	Request  reqRespOpts `json:"request"`
	Response reqRespOpts `json:"response"`
}

type dumpReqOpts struct {
	Write bool `json:"write"`
	Body  bool `json:"body"`
}

type reqRespOpts struct {
	Write  bool `json:"write"`
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
