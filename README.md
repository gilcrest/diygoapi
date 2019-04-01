# go-API-basic

A RESTful API template (built with Go) - work in progress...

- The goal of this repo/API is to make an example/template of relational database-backed APIs that have characteristics needed to ensure success in a high volume environment.

## API Walkthrough

The following is an in-depth walkthrough of this repo as well as the module dependencies that are called within. This walkthrough has a stupid amount of detail. I know I'll lose the TL;DR crowd, but, I think, for some, this might be helpful. Sections:

- [Errors](#Errors)
- [API/Handlers/Main](#Main-API-Module)

### Errors

Before even getting into the full walkthrough, I wanted to review the [errors module](https://github.com/gilcrest/errors) and the approach to error handling. The errors module is basically a carve out of the error handling used in the [upspin library](https://github.com/upspin/upspin/tree/master/errors) with some tweaks and additions I made for my own needs. Rob Pike has a [fantastic post](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) about errors and the Upspin implementation. I've taken that and added my own twist.

My general idea for error handling throughout this API and dependent modules is to always raise an error using the errors.E function as seen in this simple error handle below. Errors.E is neat - you can pass in any one of a number of approved types and the function helps form the error. In all error cases, I pass the errors.Op as the errors.E function helps build a pseudo stack trace for the error as it goes up through the code. Here's a snippet showing a typical, simple example of using the errors.E function.

```go
func NewServer(name env.Name, lvl zerolog.Level) (*Server, error) {
    const op errors.Op = "server/NewServer"

    // call constructor for Env struct from env module
    env, err := env.NewEnv(name, lvl)
    if err != nil {
        return nil, errors.E(op, err)
    }
```

The following snippet shows a more robust validation example. In it, you'll notice that if you need to define your own error quickly, you can just use a string and insert that into an error as well.

```go
func (m *Movie) validate() error {
    const op errors.Op = "movie/Movie.validate"

    switch {
    case m.Title == "":
        return errors.E(op, errors.Validation, errors.Parameter("Title"), errors.MissingField("Title"))
    case m.Year < 1878:
        return errors.E(op, errors.Validation, errors.Parameter("Year"), "The first film was in 1878, Year must be >= 1878")
```

In this you'll notice the errors.MissingField function used to validate missing input on fields, which comes from [this Mat Ryer post](https://medium.com/@matryer/patterns-for-decoding-and-validating-input-in-go-data-apis-152291ac7372)

```go
// MissingField is an error type that can be used when
// validating input fields that do not have a value, but should
type MissingField string

func (e MissingField) Error() string {
    return string(e) + " is required"
}
```

As stated before, as errors go up the stack from whatever depth of code they're in, Upspin captures the operation and adds that to the error string as a pseudo stack trace that is super helpful for debugging. However, I don't want this type of internal stack information exposed to end users in the response - I only want the error message. As such, just prior to shipping the response, I log the error (to capture the stack info) and call a custom function I built called `errors.RE` (**R**esponse **E**rror). This function effectively strips the stack information and just sends the original error message along with whatever http status code you select as well as whatever errors.Kind, Code or Parameter you choose to set. An example of error handling at the highest level (from the POST handler) is below:

```go
// Call the create method of the Movie object to validate and insert the data
err = movie.Create(ctx, s.Logger, tx)
if err != nil {
    // log error
    s.Logger.Error().Err(err).Msg("")
    // Type assertion is used - all errors should be an *errors.Error type
    // Use Kind, Param, Code and Error from lower level errors to populate RE (Response Error)
    if e, ok := err.(*errors.Error); ok {
        err := errors.RE(http.StatusBadRequest, e.Kind, e.Param, e.Code, err)
        errors.HTTPError(w, err)
        return
    }

    // if falls through type assertion, then serve an unanticipated error
    err := errors.RE(http.StatusInternalServerError, errors.Unanticipated)
    errors.HTTPError(w, err)
    return
}
```

The response body will look something like this:

```json
{
    "error": {
        "kind": "input_validation_error",
        "param": "Year",
        "message": "The first film was in 1878, Year must be >= 1878"
    }
}
```

### Main API Module

The [Main API/server layer module](https://github.com/gilcrest/go-api-basic) is the starting point and as such has the main package/function within the cmd directory. In it, I'm checking for both a log level and environment command line flags and running them through 2 simple functions (`logLevel` and `envName`) to get the correct string value given the flag input. I'm using [zerolog](https://github.com/rs/zerolog) throughout my modules as the logger.

```go
func main() {

    // loglvl flag allows for setting logging level, e.g. to run the server
    // with level set to debug, it'd be: ./server loglvl=debug
    loglvlFlag := flag.String("loglvl", "error", "sets log level")

    // env flag allows for setting environment, e.g. Production, QA, etc.
    // example: env=dev, env=qa, env=stg, env=prod
    envFlag := flag.String("env", "dev", "sets log level")

    flag.Parse()

    // get appropriate zerolog.Level based on flag
    lvl := logLevel(loglvlFlag)
    log.Log().Msgf("Logging Level set to %s", lvl)

    // get appropriate env.Name based on flag
    eName := envName(envFlag)
    log.Log().Msgf("Environment set to %s", eName)
```

Next in the `main`, I'm calling the `NewServer` function from the `server` package to construct a new server using the environment name and logging level.

```go
    // call constructor for Server struct
    server, err := server.NewServer(eName, lvl)
    if err != nil {
        log.Fatal().Err(err).Msg("")
    }
```

The server's multiplex router is Gorilla, which is registered as the handler for http.Handle.

```go
    // handle all requests with the Gorilla router
    http.Handle("/", server.Router)
```

Finally, http.ListenAndServe is run to truly start the server and listen for incoming requests.

```go
    // ListenAndServe on port 8080, not specifying a particular IP address
    // for this particular implementation
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal().Err(err).Msg("")
    }
```

Let's dig into the `server.NewServer` constructor function from above. In the `server` package, there is also a `Server` struct, which is composed of a pointer to the `Env` struct from the [env module](https://github.com/gilcrest/env), as below:

```go
// Server struct contains the environment (env.Env) and additional methods
// for running our HTTP server
type Server struct {
    *env.Env
}
```

This struct uses [type embedding](https://golang.org/doc/effective_go.html#embedding), which allows the `Server` struct to take on all the properties of the `Env` struct from the `env module` (things like database setup, logger, multiplexer, etc.), but also allows for extending the struct with API-specific methods for routing logic and handlers. The `Env` struct in the env module has the following structure:

```go
// Env struct stores common environment related items
type Env struct {
    // Environment Name (e.g. Production, QA, etc.)
    Name Name
    // multiplex router
    Router *mux.Router
    // Datastore struct containing AppDB (PostgreSQL),
    // LogDb (PostgreSQL) and CacheDB (Redis)
    DS *datastore.Datastore
    // Logger
    Logger zerolog.Logger
}
```

Within the `server.NewServer` constructor function, I'm calling the `NewEnv` function of the `env module`. The `env.NewEnv` constructor function does all the database setup, starts the logger as well as initializes the gorilla multiplexer. Fore more information on what's happening inside the `env.NewEnv` function, reference the [env module Readme](https://github.com/gilcrest/env)

```go
// NewServer is a constructor for the Server struct
// Sets up the struct and registers routes
func NewServer(name env.Name, lvl zerolog.Level) (*Server, error) {
    const op errors.Op = "server/NewServer"

    // call constructor for Env struct from env module
    env, err := env.NewEnv(name, lvl)
    if err != nil {
        return nil, errors.E(op, err)
    }
```

After getting `env.Env` back, I embed it in the Server struct

```go
    // Use type embedding to make env.Env struct part of Server struct
    server := &Server{env}
```

Finally, I call the `server.routes` method and return the server.

```go
    // routes registers handlers to the Server router
    err = server.routes()
    if err != nil {
        return nil, errors.E(op, err)
    }

    return server, nil

```

Inside the `server.routes` method, first I pull out the app database from the server struct to pass into my `servertoken` handler.

```go
// routes registers handlers to the router
func (s *Server) routes() error {
    const op errors.Op = "server/Server.routes"

    // Get App Database for token authentication
    appdb, err := s.DS.DB(datastore.AppDB)
    if err != nil {
        return errors.E(op, err)
    }
```

Next, the URL path and handlers are register to the router embedded in the server (s).

```go
    s.Router.Handle("/v1/movie",
        alice.New(
            s.handleStdResponseHeader,
            servertoken.Handler(s.Logger, appdb)).
            ThenFunc(s.handlePost())).
        Methods("POST").
        Headers("Content-Type", "application/json")
```

I am using [my own fork](https://github.com/gilcrest/alice) of [Justinas Stankevičius' alice library](https://github.com/justinas/alice) as a module to make middleware chaining easier. Hopefully the original alice library will enable modules and I'll go back, but until then I'll keep my own fork as it has properly setup modules files.

The first middleware in the chain above `s.handleStdResponseHeader` simply adds standard response headers - right now, it's just the Content-Type:application/json header, but it's an easy place to other headers one may deem standard.

The second middleware in the chain above, `servertoken.Handler` is an http handler from my [servertoken module](https://github.com/gilcrest/servertoken).