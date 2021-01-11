# go-API-basic

A RESTful API template (built with Go)

The goal of this project is to make an example/template of a relational database-backed REST HTTP API that has characteristics needed to ensure success in a high volume environment. I'm gearing this towards beginners, as I struggled with a lot of this over the past couple of years and would like to help others getting started.

[![Go Reference](https://pkg.go.dev/badge/github.com/gilcrest/go-api-basic.svg)](https://pkg.go.dev/github.com/gilcrest/go-api-basic)

## API Walkthrough

The following is an in-depth walkthrough of this project. This walkthrough has a lot of detail. This is a demo API, so the "business" intent of it is to support basic CRUD (**C**reate, **R**ead, **U**pdate, **D**elete) operations for a movie database.

## Minimum Requirements

You need to have Go and PostgreSQL installed in order to run these APIs.

### Database Setup

#### Local DB Setup

After you've installed PostgreSQL locally, the [demo_ddl.sql](https://github.com/gilcrest/go-api-basic/blob/master/demo.ddl) script (*DDL = **D**ata **D**efinition **L**anguage*) located in the root directory needs to be run, however, there are some things to know. At the highest level, PostgreSQL has the concept of databases, separate from schemas. In my script, the first statement creates a database called `go_api_basic` - this is of course optional and you can use the default postgres database or your user database or whatever you prefer. When connecting later, you'll set the database to whatever is your preference. Depending on what PostgreSQL IDE you're running the DDL in, you'll likely need to stop after this first statement, switch to this database, and then continue to run the remainder of the DDL statements. These statements create a schema (`demo`) within the database, one table (`demo.movie`) and one function (`demo.create_movie`) used on create/insert.

```sql
create database go_api_basic
    with owner postgres;
```

 In addition, [environment variables](https://en.wikipedia.org/wiki/Environment_variable) need to be in place for the database.

#### Database Connection Environment Variables

To run the app, the following environment variables need to be set:

##### PostgreSQL

```bash
export PG_APP_DBNAME="go_api_basic"
export PG_APP_USERNAME="postgres"
export PG_APP_PASSWORD=""
export PG_APP_HOST="localhost"
export PG_APP_PORT="5432"
```

You can set these however you like (permanently in something like .bash_profile if on a mac, etc. - see some notes [here](https://gist.github.com/gilcrest/d5981b873d1e2fc9646602eedd384ba6#environment-variables)), but my preferred way is to run a bash script to set the environment variables to whichever environment I'm connecting to temporarily for the current shell environment. I have included an example script file (`setlocalEnvVars.sh`) in the /scripts directory. The below statements assume you're running the command from the project root directory.

In order to set the environment variables using this script, you'll need to set the script to executable:

```bash
chmod +x ./scripts/setlocalEnvVars.sh
```

Then execute the file in the current shell environment:

```bash
source ./scripts/setlocalEnvVars.sh
```

## Installation

TL;DR - just show me how to install and run the code. Fork or clone the code.

```bash
git clone https://github.com/gilcrest/go-api-basic.git
```

To validate your installation ensure you've got connectivity to the database, do the following:

Build the code from the root directory

```bash
go build -o server
```

> This sends the output of `go build` to a file called `server` in the same directory.

Execute the file

```bash
./server -loglvl=debug
```

You should see something similar to the following:

```bash
$ ./server -loglvl=debug
{"level":"info","time":1608170937,"severity":"INFO","message":"logging level set to debug"}
{"level":"info","time":1608170937,"severity":"INFO","message":"sql database opened for localhost on port 5432"}
{"level":"info","time":1608170937,"severity":"INFO","message":"sql database Ping returned successfully"}
{"level":"info","time":1608170937,"severity":"INFO","message":"database version: PostgreSQL 12.5 on x86_64-apple-darwin16.7.0, compiled by Apple LLVM version 8.1.0 (clang-802.0.42), 64-bit"}
{"level":"info","time":1608170937,"severity":"INFO","message":"current database user: postgres"}
{"level":"info","time":1608170937,"severity":"INFO","message":"current database: go_api_basic"}
```

### Ping (unauthenticated)

The easiest api to interact with is the `ping` service. The idea of the service is a simple health check that returns a series of flags denoting health of the system (queue depths, database up boolean, etc.). For right now, the only thing it checks is if the database is up and pingable. I have left this service unauthenticated so there's at least one service that you can get to without having to have an authentication token, but in actuality, I would typically have every service behind a security token.

Use cURL GET request to call `ping`:

```bash
curl -v --location --request GET 'http://127.0.0.1:8080/api/v1/ping'
```

The response looks like:

```bash
{
    "path": "/api/v1/ping",
    "request_id": "bvfklkdnf4q0afpuo30g",
    "data": {
        "db_up": true
    }
}
```

## Authentication and Authorization

The remainder of requests require authentication. I have chosen to use [Google's Oauth2 solution](https://developers.google.com/identity/protocols/oauth2/web-server) for these APIs. In order to use Google's Oauth2, you need to setup a Client ID and Client Secret and obtain an access token. The instructions [here](https://developers.google.com/identity/protocols/oauth2) are great. I recommend the [Google Oauth2 Playground](https://developers.google.com/oauthplayground/) once you get setup to be able to easily get fresh access tokens.

Once a user has authenticated through this flow, all calls to services (other than `ping`) require that the Google access token be sent as a `Bearer` token in the `Authorization` header.

- If there is no token present, an HTTP 401 (Unauthorized) response will be sent and the response body will be empty.
- If a token is properly sent, the Google API is used to validate the token. If the token is invalid, an HTTP 401 (Unauthorized) response will be sent and the response body will be empty.
- If the token is valid, Google will respond with information about the user. The user's email will be used as their username as well as for authorization that it has been granted access to the API. If the user is not authorized to use the API, an HTTP 403 (Forbidden) response will be sent and the response body will be empty. The authorization is currently hard-coded to allow for one email. Add your email at `/domain/auth/auth.go` in the Authorize function for testing. This is definitely not a production-ready way to do authorization. I will eventually switch to some [ACL](https://en.wikipedia.org/wiki/Access-control_list) or [RBAC](https://en.wikipedia.org/wiki/Role-based_access_control) library when I have time to research those, but for now, this works.

So long as you've got a valid token and are properly setup in the authorization function, you can then execute all four operations (create, read, update, delete) using cURL.

### cURL Commands to Call API

**Create** - use the `POST` HTTP verb at `/api/v1/movies`:

```bash
curl -v --location --request POST 'http://127.0.0.1:8080/api/v1/movies' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer ya29.a0AfH6SMCLdKNT34kqZt3RhMAm4movdW4jbnb1qk8s1yOhTW6IT6r6TfddWtrYWDGrQcgSUhBiH4NOGviBE-ZBDVGb-zfDsfApOSe5tGhq_vx_v-pjKUo5g-vfALt9l5TkkXQpZ18lD47U5HhQcmM7SpRE4VwVOw4JNbFfWAYGWuCjj5KxHti9xQ' \
--data-raw '{
    "title": "Repo Man",
    "rated": "R",
    "release_date": "1984-03-02T00:00:00Z",
    "run_time": 92,
    "director": "Alex Cox",
    "writer": "Courtney Cox"
}'
```

**Read (All Records)** - use the GET HTTP verb at `/api/v1/movies`:

```bash
curl -v --location --request GET 'http://127.0.0.1:8080/api/v1/movies' \
--header 'Authorization: Bearer ya29.a0AfH6SMCLdKNT34kqZt3RhMAm4movdW4jbnb1qk8s1yOhTW6IT6r6TfddWtrYWDGrQcgSUhBiH4NOGviBE-ZBDVGb-zfDsfApOSe5tGhq_vx_v-pjKUo5g-vfALt9l5TkkXQpZ18lD47U5HhQcmM7SpRE4VwVOw4JNbFfWAYGWuCjj5KxHti9xQ' \
--data-raw ''
```

**Read (Single Record)** - use the GET HTTP verb at `/api/v1/movies/:extl_id` with the movie "external ID" from the create (POST) as the unique identifier in the URL. I try to never expose primary keys, so I use something like an external id as an alternative key.

```bash
curl -v --location --request GET 'http://127.0.0.1:8080/api/v1/movies/BDylwy3BnPazC4Casn5M' \
--header 'Authorization: Bearer ya29.a0AfH6SMCLdKNT34kqZt3RhMAm4movdW4jbnb1qk8s1yOhTW6IT6r6TfddWtrYWDGrQcgSUhBiH4NOGviBE-ZBDVGb-zfDsfApOSe5tGhq_vx_v-pjKUo5g-vfALt9l5TkkXQpZ18lD47U5HhQcmM7SpRE4VwVOw4JNbFfWAYGWuCjj5KxHti9xQ' \
--data-raw ''
```

**Update** - use the PUT HTTP verb at `/api/v1/movies/:extl_id` with the movie "external ID" from the create (POST) as the unique identifier in the URL.

```bash
curl --location --request PUT 'http://127.0.0.1:8080/api/v1/movies/BDylwy3BnPazC4Casn5M' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer ya29.a0AfH6SMCLdKNT34kqZt3RhMAm4movdW4jbnb1qk8s1yOhTW6IT6r6TfddWtrYWDGrQcgSUhBiH4NOGviBE-ZBDVGb-zfDsfApOSe5tGhq_vx_v-pjKUo5g-vfALt9l5TkkXQpZ18lD47U5HhQcmM7SpRE4VwVOw4JNbFfWAYGWuCjj5KxHti9xQ' \
--data-raw '{
    "title": "Repo Man",
    "rated": "R",
    "release_date": "1984-03-02T00:00:00Z",
    "run_time": 92,
    "director": "Alex Cox",
    "writer": "Alex Cox"
}'
```

**Delete** - use the DELETE HTTP verb at `/api/v1/movies/:extl_id` with the movie "external ID" from the create (POST) as the unique identifier in the URL.

```bash
curl --location --request DELETE 'http://127.0.0.1:8080/api/v1/movies/BDylwy3BnPazC4Casn5M' \
--header 'Authorization: Bearer ya29.a0AfH6SMCLdKNT34kqZt3RhMAm4movdW4jbnb1qk8s1yOhTW6IT6r6TfddWtrYWDGrQcgSUhBiH4NOGviBE-ZBDVGb-zfDsfApOSe5tGhq_vx_v-pjKUo5g-vfALt9l5TkkXQpZ18lD47U5HhQcmM7SpRE4VwVOw4JNbFfWAYGWuCjj5KxHti9xQ'
```

## Project Walkthrough

### Errors

Handling errors is really important in Go. Errors are first class citizens and there are many different approaches for handling them. Initially I started off basing my error handling almost entirely on a [blog post from Rob Pike](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) and created a carve-out from his code to meet my needs. It served me well for a long time, but found over time I wanted a way to easily get a stacktrace of the error, which led me to Dave Cheney's [https://github.com/pkg/errors](https://github.com/pkg/errors) package. I now use a combination of the two.

 Error handling throughout `go-api-basic` always creates an error using the `E` function from the `errs` package as seen below. `errs.E`, is derived from Rob Pike's package (but has been changed a lot). The `errs.E` function call is [variadic](https://en.wikipedia.org/wiki/Variadic) and can take several different types to form the custom `Error`struct.

 ```go
// Error is the type that implements the error interface.
// It contains a number of fields, each of different type.
// An Error value may leave some values unset.
type Error struct {
    // User is the username of the user attempting the operation.
    User UserName
    // Kind is the class of error, such as permission failure,
    // or "Other" if its class is unknown or irrelevant.
    Kind Kind
    // Param represents the parameter related to the error.
    Param Parameter
    // Code is a human-readable, short representation of the error
    Code Code
    // The underlying error that triggered this one, if any.
    Err error
}
```

Here is a simple example of creating an `error`:

```go
err := errs.E("seems we have an error here")
```

 When a string is sent, an error will be created using the `errors.New` function from `github.com/pkg/errors` and added to the `Err` element of the struct, which allows a stacktrace to be generated later on if need be. In the above example, `User`, `Kind`, `Param` and `Code` would all remain unset.

You can, of course, choose to set any of the custom error values that you like, for example:

```go
func (m *Movie) SetReleased(r string) (*Movie, error) {
    t, err := time.Parse(time.RFC3339, r)
    if err != nil {
        return nil, errs.E(errs.Validation,
            errs.Code("invalid_date_format"),
            errs.Parameter("release_date"),
            err)
    }
    m.Released = t
    return m, nil
}
```

Above, we used `errs.Validation` to set the `errs.Kind` as Validation. Valid error `Kind` are:

```go
// Kinds of errors.
//
// The values of the error kinds are common between both
// clients and servers. Do not reorder this list or remove
// any items since that will change their values.
// New items must be added only to the end.
const (
    Other           Kind = iota // Unclassified error. This value is not printed in the error message.
    Invalid                     // Invalid operation for this type of item.
    Permission                  // Permission denied.
    IO                          // External I/O error such as network failure.
    Exist                       // Item already exists.
    NotExist                    // Item does not exist.
    Private                     // Information withheld.
    Internal                    // Internal error or inconsistency.
    BrokenLink                  // Link target does not exist.
    Database                    // Error from database.
    Validation                  // Input validation error.
    Unanticipated               // Unanticipated error.
    InvalidRequest              // Invalid Request
    Unauthenticated             // User did not properly authenticate
    Unauthorized                // User is not authorized for the resource
)
```

`errs.Code` represents a short code to respond to the client with for error handling based on codes (if you choose to do this) and is any string you want to pass.

`errs.Parameter` represents the parameter that is being validated or has problems, etc.

In addition, instead of passing a string and creating a new error inside the `errs.E` function, I am just passing in the error received from the `time.Parse` function and inside `errs.E` the error is added to `Err` using `errors.WithStack` from the `github.com/pkg/errors` package so that the stacktrace can be obtained later if needed.

There are a few helpers in the `errs` package as well, namely the `errs.MissingField` function which can be used when validating missing input on a field. This idea comes from [this Mat Ryer post](https://medium.com/@matryer/patterns-for-decoding-and-validating-input-in-go-data-apis-152291ac7372) and is pretty handy.

Here is an example in practice:

```go
// IsValid performs validation of the struct
func (m *Movie) IsValid() error {
    switch {
    case m.Title == "":
        return errs.E(errs.Validation, errs.Parameter("title"), errs.MissingField("title"))
```

The error message for the above would read **title is required**

There is also `errs.InputUnwanted` which is meant to be used when a field is populated with a value when it is not supposed to be.

#### Error Flow

Errors at their initial point of failure should always start with `errs.E`, but as they move up the call stack, `errs.E` does not need to be used. Errors should just be passed on up, like the following:

```go
func inner() error {
    return errs.E("seems we have an error here")
}

func middle() error {
    err := inner()
    if err != nil {
        return err
    }
    return nil
}

func outer() error {
    err := middle()
    if err != nil {
        return err
    }
    return nil
}

```

In the above example, the error is created in the `inner` function - `middle` and `outer` return the error as is typical in Go.

At the top of the program flow for each service is the handler. I've structured my code so that handlers are relatively simple. We'll talk more about them later, but you'll notice in each handler, if any error occurs from any function/method calls, they are sent through the `errs.HTTPErrorResponse` function along with the `http.ResponseWriter` and a `zerolog.Logger`.

For example:

```go
// Send the request to the controller
// Receive a response or error in return
response, err := mc.CreateMovie(r)
if err != nil {
    errs.HTTPErrorResponse(w, logger, err)
    return
}
```

`errs.HTTPErrorResponse` takes the custom `Error` created by `errs.E` and writes the HTTP response body as JSON as well as logs the error, including the error stacktrace. When the above error is returned to the client using the `errs.HTTPErrorResponse` function in the each of the handlers, the response body JSON looks like the following:

```json
{
    "error": {
        "kind": "input_validation_error",
        "code": "invalid_date_format",
        "param": "release_date",
        "message": "parsing time \"1984a-03-02T00:00:00Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"a-03-02T00:00:00Z\" as \"-\""
    }
}
```

and the error log looks like (I cut off parts of the stack for brevity):

```json
{
    "level": "error",
    "ip": "127.0.0.1",
    "user_agent": "PostmanRuntime/7.26.8",
    "request_id": "bvol0mtnf4q269hl3ra0",
    "stack": [{
        "func": "E",
        "line": "172",
        "source": "errs.go"
    }, {
        "func": "(*Movie).SetReleased",
        "line": "76",
        "source": "movie.go"
    }, {
        "func": "(*MovieController).CreateMovie",
        "line": "139",
        "source": "create.go"
    }, {
    ...
    }],
    "error": "parsing time \"1984a-03-02T00:00:00Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"a-03-02T00:00:00Z\" as \"-\"",
    "HTTPStatusCode": 400,
    "Kind": "input_validation_error",
    "Parameter": "release_date",
    "Code": "invalid_date_format",
    "time": 1609650267,
    "severity": "ERROR",
    "message": "Response Error Sent"
}
```

> Note: `E` will often be at the top of the stack as it is where the `errors.New` or `errors.WithStack` functions are being called. If you prefer not to see this, you can call `errors.New` or `errors.WithStack` as part of the `errs.E` call, for example:

```go
err := errs.E(errors.New("seems we have an error here"))
```

## 1/3/2021 - README under construction

I have taken out the remainder of the documentation for now until I complete my next goal of adding more tests to just about everything. I think adding tests will likely further shape the structure and program flow that I'm going to wait until I've completed that exercise to complete this README.
