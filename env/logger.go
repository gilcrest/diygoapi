package env

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog"
)

// HTTPLogOpts represent HTTP Logging Options
type HTTPLogOpts struct {
	Log2StdOut *log2StdOut `json:"log_json"`
	Log2DB     *log2DB     `json:"log_2DB"`
	HTTPUtil   *httputil   `json:"httputil"`
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

func newHTTPLogOpts() (*HTTPLogOpts, error) {

	raw, err := ioutil.ReadFile("../input/httpLogOpt.json")
	if err != nil {
		return nil, err
	}

	var l HTTPLogOpts
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, err
	}

	return &l, nil
}

func newLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = ""
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return lgr
}
