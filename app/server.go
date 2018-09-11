package app

import (
	"fmt"
	"os"

	"github.com/gilcrest/go-API-template/datastore"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

// Server struct stores common environment related items
type server struct {
	router *mux.Router
	ds     *datastore.Datastore
	logger zerolog.Logger
}

// NewServer initializes the Server struct
func NewServer(lvl zerolog.Level) (*server, error) {

	// setup logger
	log := newLogger(lvl)

	// open db connection pools
	dstore, err := datastore.NewDatastore()
	if err != nil {
		return nil, err
	}

	// create a new mux (multiplex) router
	rtr := mux.NewRouter()

	// send Router through subRouter function to add any standard
	// Subroutes you may want for your APIs
	rtr = newSubrouter(rtr)

	server := &server{router: rtr, ds: dstore, logger: log}

	server.routes()

	return server, nil
}

// LogErr logs the Operation (client.Lookup, etc.) as well
// as the error string and returns an error
func (s *server) LogErr(op string, str string) error {
	err := fmt.Errorf("%s: %s", op, str)
	s.logger.Error().Err(err).Msg("")
	return err
}

// newSubrouter adds any subRouters that you'd like to have as part of
// every request, i.e. I always want to be sure that every request has
// "/api" as part of it's path prefix without having to put it into
// every handle path in my various routing functions
func newSubrouter(rtr *mux.Router) *mux.Router {
	sRtr := rtr.PathPrefix("/api").Subrouter()
	return sRtr
}

// newLogger sets up
func newLogger(lvl zerolog.Level) zerolog.Logger {
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(lvl)
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return lgr
}

func (s *server) Router() *mux.Router {
	return s.router
}
