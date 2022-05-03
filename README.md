# DIY Go API

v0.43.0 - 5/3/2022 - I have refactored large swaths of the app recently around the way I was doing db rollbacks... I have started to implement a DIY RBAC, but am still wrapping up some of the administration services (update role, update permission, etc.). I am going to update the module name with the next commit to be consistent with the repo name.

Things are still coming back together, a ton of work recently. **The docs below are out of date in many cases**. Hoping to have everything sorted soon, but if you look at the code, you can start to see the structure. I am currently building a simple RBAC implementation as part of this as well. After that is completed, README documentation will be my focus. I will get this updated!

--------

A RESTful API template (built with Go)

The goal of this project is to be an example of a relational database-backed REST HTTP API that has characteristics needed to ensure success in a high volume environment. I struggled a lot with parsing all of the different ideas people have for package layouts over the past few years and would like to help others who may be having a similar struggle. I'd like to see Go as accessible as possible for everyone. If you have any questions or would like help, open an issue or send me a note - I'm happy to help! Also, if you have disagree or have suggestions about this repo, please do the same, I really enjoy getting both positive and negative feedback.

[![Go Reference](https://pkg.go.dev/badge/github.com/gilcrest/go-api-basic.svg)](https://pkg.go.dev/github.com/gilcrest/go-api-basic) [![Go Report Card](https://goreportcard.com/badge/github.com/gilcrest/diy-go-api)](https://goreportcard.com/report/github.com/gilcrest/diy-go-api)

## API Walkthrough

The following is an in-depth walkthrough of this project. This walkthrough has a lot of detail. This is a demo API, so the "business" intent of it is to support basic CRUD (**C**reate, **R**ead, **U**pdate, **D**elete) operations for a movie database.

## Minimum Requirements

Go and PostgreSQL are required in order to run these APIs. In addition, several database objects must be created (see [Database Objects Setup](#database-objects-setup) below)

## Table of Contents

- [Getting Started](#getting-started)
  - [Database Objects Setup](#database-objects-setup)
  - [Program Execution](#program-execution)
    - [Command Line Flags](#command-line-flags)
    - [Environment Setup](#environment-setup)
      - [Database Connection Environment Variables](database-connection-environment-variables)
    - [Run the Binary](#run-the-binary)
  - [Ping](#ping)
  - [Authentication and Authorization](#authentication-and-authorization)
  - [cURL Commands to Call Services](#curl-commands-to-call-services)
  - [Project Walkthrough](#project-walkthrough)
    - [Errors](#errors)
    - [Logging](#logging)

---

## Getting Started

### Database Objects Setup

Assuming PostgreSQL is installed locally, the [demo_ddl.sql](https://github.com/gilcrest/go-api-basic/blob/master/demo.ddl) script (*DDL = **D**ata **D**efinition **L**anguage*) located in the `/scripts/ddl` directory needs to be run, however, there are some things to know. At the highest level, PostgreSQL has the concept of databases, separate from schemas. A database is a container of other objects (tables, views, functions, indexes, etc.). There is no limit no the number of databases inside a PostgreSQL server.

In the DDL script, the first statement creates a database called `go_api_basic`:

```sql
create database go_api_basic
    with owner postgres;
```

> Using this database is optional; you can use the default `postgres` database or your user database or whatever you prefer. When connecting later, you can set the database to whatever is your preference. If you do choose to create a separate database like this, depending on what PostgreSQL IDE you're running the DDL in, you'll likely need to stop after this first statement, switch to this database, and then continue to run the remainder of the DDL statements.

The remainder of the statements create a schema (`demo`) within the database:

```sql
create schema demo;
```

one table (`demo.movie`):

```sql
create table demo.movie
(
    movie_id uuid not null
        constraint movie_pk
            primary key,
    extl_id varchar(250) not null,
    title varchar(1000) not null,
    rated varchar(10),
    released date,
    run_time integer,
    director varchar(1000),
    writer varchar(1000),
    create_username varchar,
    create_timestamp timestamp with time zone,
    update_username varchar,
    update_timestamp timestamp with time zone
);

alter table demo.movie owner to postgres;

create unique index movie_extl_id_uindex
    on demo.movie (extl_id);
```

and one function (`demo.create_movie`) used on create/insert:

```sql
create function demo.create_movie(p_id uuid, p_extl_id character varying, p_title character varying, p_rated character varying, p_released date, p_run_time integer, p_director character varying, p_writer character varying, p_create_client_id uuid, p_create_username character varying)
    returns TABLE(o_create_timestamp timestamp without time zone, o_update_timestamp timestamp without time zone)
    language plpgsql
as
$$
DECLARE
    v_dml_timestamp TIMESTAMP;
    v_create_timestamp timestamp;
    v_update_timestamp timestamp;
BEGIN

    v_dml_timestamp := now() at time zone 'utc';

    INSERT INTO demo.movie (movie_id,
                            extl_id,
                            title,
                            rated,
                            released,
                            run_time,
                            director,
                            writer,
                            create_username,
                            create_timestamp,
                            update_username,
                            update_timestamp)
    VALUES (p_id,
            p_extl_id,
            p_title,
            p_rated,
            p_released,
            p_run_time,
            p_director,
            p_writer,
            p_create_username,
            v_dml_timestamp,
            p_create_username,
            v_dml_timestamp)
    RETURNING create_timestamp, update_timestamp
        into v_create_timestamp, v_update_timestamp;

    o_create_timestamp := v_create_timestamp;
    o_update_timestamp := v_update_timestamp;

    RETURN NEXT;

END;

$$;

alter function demo.create_movie(uuid, varchar, varchar, varchar, date, integer, varchar, varchar, uuid, varchar) owner to postgres;
```

### Program Execution

TL;DR - just show me how to install and run the code. Fork or clone the code.

```bash
git clone https://github.com/gilcrest/go-api-basic.git
```

To validate your installation and ensure you've got connectivity to the database, do the following:

Build the code from the program root directory

```bash
go build -o server
```

> This sends the output of `go build` to a binary file called `server` in the same directory.

#### Command Line Flags

When running the program binary, a number flags can be passed. The [ff](https://github.com/peterbourgon/ff) library from [Peter Bourgon](https://peter.bourgon.org) is used to parse the flags. If your preference is to set configuration with [environment variables](https://en.wikipedia.org/wiki/Environment_variable), that is possible as well. Flags take precedence, so if a flag is passed, that will be used. A PostgreSQL database connection is required. If there is no flag set, then the program checks for a matching environment variable. If neither are found, the flag's default value will be used and, depending on the flag, may result in a database connection error.

| Flag Name       | Description | Environment Variable | Default |
| --------------- | ----------- | -------------------- | ------- |
| port            | Port the server will listen on | PORT | 8080|
| log-level       | zerolog logging level (debug, info, etc.) | LOG_LEVEL | debug |
| log-level-min   | sets the minimum accepted logging level | LOG_LEVEL_MIN | debug |
| log-error-stack | If true, log full error stacktrace, else just log error | LOG_ERROR_STACK | false |
| db-host         | The host name of the database server. | DB_HOST | |
| db-port         | The port number the database server is listening on.| DB_PORT | 5432 |
| db-name         | The database name. | DB_NAME | |
| db-user         | PostgreSQL™ user name to connect as. | DB_USER | |
| db-password     | Password to be used if the server demands password authentication. | DB_PASSWORD | |

#### Environment Setup

If you choose to use [environment variables](https://en.wikipedia.org/wiki/Environment_variable) instead of flags for connecting to the database, you can set these however you like (permanently in something like .`bash_profile` if on a mac, etc. - some notes [here](https://gist.github.com/gilcrest/d5981b873d1e2fc9646602eedd384ba6#environment-variables)), but my preferred way is to run a bash script to set environment variables temporarily for the current shell environment. I have included an example script file (`setlocalEnvVars.sh`) in the `/scripts/ddl` directory. The below statements assume you're running the command from the project root directory.

In order to set the environment variables using this script, first, in the file, set the environment variable values to whatever is appropriate for your environment:

##### Database Connection Environment Variables

```bash
export DB_NAME="go_api_basic"
export DB_USER="postgres"
export DB_PASSWORD=""
export DB_HOST="localhost"
export DB_PORT="5432"
```

Next, you'll need to set the script to executable

```bash
chmod +x ./scripts/setlocalEnvVars.sh
```

Finally, execute the file in the current shell environment:

```bash
source ./scripts/ddl/setlocalEnvVars.sh
```

#### Run the Binary

```bash
./server -log-level=debug -db-host=localhost -db-port=5432 -db-name=go_api_basic -db-user=postgres -db-password=fakePassword
```

Upon running, you should see something similar to the following:

```bash
$ ./server -log-level=debug
{"level":"info","time":1618260160,"severity":"INFO","message":"minimum accepted logging level set to trace"}
{"level":"info","time":1618260160,"severity":"INFO","message":"logging level set to debug"}
{"level":"info","time":1618260160,"severity":"INFO","message":"log error stack global set to true"}
{"level":"info","time":1618260160,"severity":"INFO","message":"sql database opened for localhost on port 5432"}
{"level":"info","time":1618260160,"severity":"INFO","message":"sql database Ping returned successfully"}
{"level":"info","time":1618260160,"severity":"INFO","message":"database version: PostgreSQL 12.6 on x86_64-apple-darwin16.7.0, compiled by Apple LLVM version 8.1.0 (clang-802.0.42), 64-bit"}
{"level":"info","time":1618260160,"severity":"INFO","message":"current database user: postgres"}
{"level":"info","time":1618260160,"severity":"INFO","message":"current database: go_api_basic"}
```

##### Ping

With the server up and running, the easiest service to interact with is the `ping` service. This service is a simple health check that returns a series of flags denoting health of the system (queue depths, database up boolean, etc.). For right now, the only thing it checks is if the database is up and pingable. I have left this service unauthenticated so there's at least one service that you can get to without having to have an authentication token, but in actuality, I would typically have every service behind a security token.

Use [cURL](https://curl.se/) GET request to call `ping`:

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

### Authentication and Authorization

The remainder of requests require authentication. I have chosen to use [Google's Oauth2 solution](https://developers.google.com/identity/protocols/oauth2/web-server) for these APIs. To use this, you need to setup a Client ID and Client Secret and obtain an access token. The instructions [here](https://developers.google.com/identity/protocols/oauth2) are great.

After Oauth2 setup with Google, I recommend the [Google Oauth2 Playground](https://developers.google.com/oauthplayground/) to obtain fresh access tokens for testing.

Once a user has authenticated through this flow, all calls to services (other than `ping`) require that the Google access token be sent as a `Bearer` token in the `Authorization` header.

- If there is no token present, an HTTP 401 (Unauthorized) response will be sent and the response body will be empty.
- If a token is properly sent, the Google API is used to validate the token. If the token is invalid, an HTTP 401 (Unauthorized) response will be sent and the response body will be empty.
- If the token is valid, Google will respond with information about the user. The user's email will be used as their username as well as for authorization that it has been granted access to the API. If the user is not authorized to use the API, an HTTP 403 (Forbidden) response will be sent and the response body will be empty. The authorization is currently hard-coded to allow for one email. Add your email at `/domain/auth/auth.go` in the `Authorize` method of the `Authorizer` struct for testing. This is definitely not a production-ready way to do authorization. I will eventually switch to some [ACL](https://en.wikipedia.org/wiki/Access-control_list) or [RBAC](https://en.wikipedia.org/wiki/Role-based_access_control) library when I have time to research those, but for now, this works.

So long as you've got a valid token and are properly setup in the authorization function, you can then execute all four operations (create, read, update, delete) using cURL.

### cURL Commands to Call Services

**Create** - use the `POST` HTTP verb at `/api/v1/movies`:

```bash
curl -v --location --request POST 'http://127.0.0.1:8080/api/v1/movies' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
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
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
--data-raw ''
```

**Read (Single Record)** - use the GET HTTP verb at `/api/v1/movies/:extl_id` with the movie "external ID" from the create (POST) as the unique identifier in the URL. I try to never expose primary keys, so I use something like an external id as an alternative key.

```bash
curl -v --location --request GET 'http://127.0.0.1:8080/api/v1/movies/BDylwy3BnPazC4Casn5M' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
--data-raw ''
```

**Update** - use the PUT HTTP verb at `/api/v1/movies/:extl_id` with the movie "external ID" from the create (POST) as the unique identifier in the URL.

```bash
curl --location --request PUT 'http://127.0.0.1:8080/api/v1/movies/BDylwy3BnPazC4Casn5M' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
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
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>'
```

## Project Walkthrough

### Errors

Handling errors is really important in Go. Errors are first class citizens and there are many different approaches for handling them. Initially I started off basing my error handling almost entirely on a [blog post from Rob Pike](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) and created a carve-out from his code to meet my needs. It served me well for a long time, but found over time I wanted a way to easily get a stacktrace of the error, which led me to Dave Cheney's [https://github.com/pkg/errors](https://github.com/pkg/errors) package. I now use a combination of the two.

#### Error Requirements

My requirements for REST API error handling are the following:

- Requests for users who are *not* properly ***authenticated*** should return a `401 Unauthorized` error with a `WWW-Authenticate` response header and an empty response body.
- Requests for users who are authenticated, but do not have permission to access the resource, should return a `403 Forbidden` error with an empty response body.
- All requests which are due to a client error (invalid data, malformed JSON, etc.) should return a `400 Bad Request` and a response body which looks similar to the following:

```json
{
    "error": {
        "kind": "input_validation_error",
        "param": "director",
        "message": "director is required"
    }
}
```

- All requests which incur errors as a result of an internal server or database error should return a `500 Internal Server Error` and not leak any information about the database or internal systems to the client. These errors should return a response body which looks like the following:

```json
{
    "error": {
        "kind": "internal_error",
        "message": "internal server error - please contact support"
    }
}
```

All errors should return a `Request-Id` response header with a unique request id that can be used for debugging to find the corresponding error in logs.

#### Error Implementation

All errors should be raised using custom errors from the [domain/errs](https://github.com/gilcrest/go-api-basic/tree/main/domain/errs) package. The three custom errors correspond directly to the requirements above.

##### Typical Errors

Typical errors raised throughout `go-api-basic` are the custom `errs.Error`, which look like:

 ```go
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

This custom error type is raised using the `E` function from the [domain/errs](https://github.com/gilcrest/go-api-basic/tree/main/domain/errs) package. `errs.E` is taken from Rob Pike's [upspin errors package](https://github.com/upspin/upspin/tree/master/errors) (but has been changed based on my requirements). The `errs.E` function call is [variadic](https://en.wikipedia.org/wiki/Variadic) and can take several different types to form the custom `errs.Error` struct.

Here is a simple example of creating an `error` using `errs.E`:

```go
err := errs.E("seems we have an error here")
```

When a string is sent, an error will be created using the `errors.New` function from `github.com/pkg/errors` and added to the `Err` element of the struct, which allows retrieval of the error stacktrace later on. In the above example, `User`, `Kind`, `Param` and `Code` would all remain unset.

You can set any of these custom `errs.Error` fields that you like, for example:

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

Above, we used `errs.Validation` to set the `errs.Kind` as `Validation`. Valid error `Kind` are:

```go
const (
    Other           Kind = iota // Unclassified error. This value is not printed in the error message.
    Invalid                     // Invalid operation for this type of item.
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
)
```

`errs.Code` represents a short code to respond to the client with for error handling based on codes (if you choose to do this) and is any string you want to pass.

`errs.Parameter` represents the parameter that is being validated or has problems, etc.

> Note in the above example, instead of passing a string and creating a new error inside the `errs.E` function, I am directly passing the error returned by the `time.Parse` function to `errs.E`. The error is then added to the `Err` field using `errors.WithStack` from the `github.com/pkg/errors` package, which enables stacktrace retrieval later.

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

###### Typical Error Flow

As errors created with `errs.E` move up the call stack, they can just be returned, like the following:

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

> In the above example, the error is created in the `inner` function - `middle` and `outer` return the error as is typical in Go.

You can add additional context fields (`errs.Code`, `errs.Parameter`, `errs.Kind`) as the error moves up the stack, however, I try to add as much context as possible at the point of error origin and only do this in rare cases.

##### Handler Flow

At the top of the program flow for each service is the app service handler (for example, [Server.handleMovieCreate](https://github.com/gilcrest/go-api-basic/blob/main/app/handlers.go)). In this handler, any error returned from any function or method is sent through the `errs.HTTPErrorResponse` function along with the `http.ResponseWriter` and a `zerolog.Logger`.

For example:

```go
response, err := s.CreateMovieService.Create(r.Context(), rb, u)
if err != nil {
    errs.HTTPErrorResponse(w, logger, err)
    return
}
```

`errs.HTTPErrorResponse` takes the custom error (`errs.Error`, `errs.Unauthenticated` or `errs.UnauthorizedError`), writes the response to the given `http.ResponseWriter` and logs the error using the given `zerolog.Logger`.

> `return` must be called immediately after `errs.HTTPErrorResponse` to return the error to the client.

##### Typical Error Response

For the `errs.Error` type, `errs.HTTPErrorResponse` writes the HTTP response body as JSON using the `errs.ErrResponse` struct.

```go
// ErrResponse is used as the Response Body
type ErrResponse struct {
    Error ServiceError `json:"error"`
}

// ServiceError has fields for Service errors. All fields with no data will
// be omitted
type ServiceError struct {
    Kind    string `json:"kind,omitempty"`
    Code    string `json:"code,omitempty"`
    Param   string `json:"param,omitempty"`
    Message string `json:"message,omitempty"`
}
```

When the error is returned to the client, the response body JSON looks like the following:

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

In addition, the error is logged. If `zerolog.ErrorStackMarshaler` is set to log error stacks (more about this below), the logger will log the full error stack, which can be super helpful when trying to identify issues.

The error log will look like the following (*I cut off parts of the stack for brevity*):

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

> Note: `E` will usually be at the top of the stack as it is where the `errors.New` or `errors.WithStack` functions are being called.

##### Internal or Database Error Response

There is logic within `errs.HTTPErrorResponse` to return a different response body if the `errs.Kind` is `Internal` or `Database`. As per the requirements, we should not leak the error message or any internal stack, etc. when an internal or database error occurs. If an error comes through and is an `errs.Error` with either of these error `Kind` or is unknown error type in any way, the response will look like the following:

```json
{
    "error": {
        "kind": "internal_error",
        "message": "internal server error - please contact support"
    }
}
```

---

#### Unauthenticated Errors

```go
type UnauthenticatedError struct {
    // WWWAuthenticateRealm is a description of the protected area.
    // If no realm is specified, "DefaultRealm" will be used as realm
    WWWAuthenticateRealm string

    // The underlying error that triggered this one, if any.
    Err error
}
```

The [spec](https://tools.ietf.org/html/rfc7235#section-3.1) for `401 Unauthorized` calls for a `WWW-Authenticate` response header along with a `realm`. The realm should be set when creating an Unauthenticated error. The `errs.NewUnauthenticatedError` function initializes an `UnauthenticatedError`.

> I generally like to follow the Go idiom for brevity in all things as much as possible, but for `Unauthenticated` vs. `Unauthorized` errors, it's confusing enough as it is already, I don't take any shortcuts.

```go
func NewUnauthenticatedError(realm string, err error) *UnauthenticatedError {
    return &UnauthenticatedError{WWWAuthenticateRealm: realm, Err: err}
}
```

##### Unauthenticated Error Flow

The `errs.Unauthenticated` error should only be raised at points of authentication as part of a middleware handler. I will get into application flow in detail later, but authentication for `go-api-basic` happens in middleware handlers prior to calling the app handler for the given route.

- The `WWW-Authenticate` *realm* is set to the request context using the `defaultRealmHandler` middleware in the [app package](https://github.com/gilcrest/go-api-basic/blob/main/app/middleware.go) prior to attempting authentication.
- Next, the Oauth2 access token is retrieved from the `Authorization` http header using the `accessTokenHandler` middleware. There are several access token validations in this middleware, if any are not successful, the `errs.Unauthenticated` error is returned using the realm set to the request context.
- Finally, if the access token is successfully retrieved, it is then converted to a `User` via the `GoogleAccessTokenConverter.Convert` method in the `gateway/authgateway` package. This method sends an outbound request to Google using their API; if any errors are returned, an `errs.Unauthenticated` error is returned.

> In general, I do not like to use `context.Context`, however, it is used in `go-api-basic` to pass values between middlewares. The `WWW-Authenticate` *realm*, the Oauth2 access token and the calling user after authentication, all of which are `request-scoped` values, are all set to the request `context.Context`.

##### Unauthenticated Error Response

Per requirements, `go-api-basic` does not return a response body when returning an **Unauthenticated** error. The error response from [cURL](https://curl.se/) looks like the following:

```bash
HTTP/1.1 401 Unauthorized
Request-Id: c30hkvua0brkj8qhk3e0
Www-Authenticate: Bearer realm="go-api-basic"
Date: Wed, 09 Jun 2021 19:46:07 GMT
Content-Length: 0
```

---

#### Unauthorized Errors

```go
type UnauthorizedError struct {
    // The underlying error that triggered this one, if any.
    Err error
}
```

The `errs.NewUnauthorizedError` function initializes an `UnauthorizedError`.

##### Unauthorized Error Flow

The `errs.Unauthorized` error is raised when there is a permission issue for a user when attempting to access a resource. Currently, `go-api-basic`'s placeholder authorization implementation `Authorizer.Authorize` in the [domain/auth](https://github.com/gilcrest/go-api-basic/blob/main/domain/auth/auth.go) package performs rudimentary checks that a user has access to a resource. If the user does not have access, the `errs.Unauthorized` error is returned.

Per requirements, `go-api-basic` does not return a response body when returning an **Unauthorized** error. The error response from [cURL](https://curl.se/) looks like the following:

```bash
HTTP/1.1 403 Forbidden
Request-Id: c30hp2ma0brkj8qhk3f0
Date: Wed, 09 Jun 2021 19:54:50 GMT
Content-Length: 0
```

### Logging

`go-api-basic` uses the [zerolog](https://github.com/rs/zerolog) library from [Olivier Poitrey](https://github.com/rs). The mechanics for using `zerolog` are straightforward and are well documented in the library's [README](https://github.com/rs/zerolog#readme). `zerolog` takes an `io.Writer` as input to create a new logger; for simplicity in `go-api-basic`, I use `os.Stdout`.

#### Setting Logger State on Startup

When starting `go-api-basic`, there are several flags which setup the logger:

| Flag Name       | Description | Environment Variable | Default |
| --------------- | ----------- | -------------------- | ------- |
| log-level       | zerolog logging level (debug, info, etc.) | LOG_LEVEL | debug |
| log-level-min   | sets the minimum accepted logging level | LOG_LEVEL_MIN | debug |
| log-error-stack | If true, log full error stacktrace, else just log error | LOG_ERROR_STACK | false |

---

> As mentioned [above](https://github.com/gilcrest/go-api-basic#command-line-flags), `go-api-basic` uses the [ff](https://github.com/peterbourgon/ff) library from [Peter Bourgon](https://peter.bourgon.org), which allows for using either flags or environment variables. Going forward, we'll assume you've chosen flags.

The `log-level` flag sets the Global logging level for your `zerolog.Logger`.

**zerolog** allows for logging at the following levels (from highest to lowest):

- panic (`zerolog.PanicLevel`, 5)
- fatal (`zerolog.FatalLevel`, 4)
- error (`zerolog.ErrorLevel`, 3)
- warn (`zerolog.WarnLevel`, 2)
- info (`zerolog.InfoLevel`, 1)
- debug (`zerolog.DebugLevel`, 0)
- trace (`zerolog.TraceLevel`, -1)

The `log-level-min` flag sets the minimum accepted logging level, which means, for example, if you set the minimum level to error, the only logs that will be sent to your chosen output will be those that are greater than or equal to error (`error`, `fatal` and `panic`).

The `log-error-stack` boolean flag tells whether to log stack traces for each error. If `true`, the `zerolog.ErrorStackMarshaler` will be set to `pkgerrors.MarshalStack` which means, for errors raised using the [github.com/pkg/errors](https://github.com/pkg/errors) package, the error stack trace will be captured and printed along with the log. All errors raised in `go-api-basic` are raised using `github.com/pkg/errors`.

After parsing the command line flags, `zerolog.Logger` is initialized in `main.go`

```go
// setup logger with appropriate defaults
lgr := logger.NewLogger(os.Stdout, minlvl, true)
```

and subsequently injected into the `app.Server` struct as a Server parameter.

```go
// initialize server configuration parameters
params := app.NewServerParams(lgr, serverDriver)

// initialize Server
s, err := app.NewServer(mr, params)
if err != nil {
    lgr.Fatal().Err(err).Msg("Error from app.NewServer")
}
```

#### Logger Setup in Handlers

The `Server.routes` method is responsible for registering routes and corresponding middleware/handlers to the Server's `gorilla/mux` router. For each route registered to the handler, upon execution, the initialized `zerolog.Logger` struct is added to the request context through the `Server.loggerChain` method.

```go
// register routes/middleware/handlers to the Server router
func (s *Server) routes() {

    // Match only POST requests at /api/v1/movies
    // with Content-Type header = application/json
    s.router.Handle(moviesV1PathRoot,
        s.loggerChain().Extend(s.ctxWithUserChain()).
            Append(s.authorizeUserHandler).
            Append(s.jsonContentTypeResponseHandler).
            ThenFunc(s.handleMovieCreate)).
        Methods(http.MethodPost).
        Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

...
```

 The `Server.loggerChain` method sets up the logger with pre-populated fields, including the request method, url, status, size, duration, remote IP, user agent, referer. A unique `Request ID` is also added to the logger, context and response headers.

```go
func (s *Server) loggerChain() alice.Chain {
    ac := alice.New(hlog.NewHandler(s.logger),
        hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
        hlog.FromRequest(r).Info().
            Str("method", r.Method).
            Stringer("url", r.URL).
            Int("status", status).
            Int("size", size).
            Dur("duration", duration).
            Msg("request logged")
        }),
        hlog.RemoteAddrHandler("remote_ip"),
        hlog.UserAgentHandler("user_agent"),
        hlog.RefererHandler("referer"),
        hlog.RequestIDHandler("request_id", "Request-Id"),
    )

    return ac
}
```

For every request, you'll get a request log that looks something like the following:

```json
{
    "level": "info",
    "remote_ip": "127.0.0.1",
    "user_agent": "PostmanRuntime/7.28.0",
    "request_id": "c3npn8ea0brt0m3scvq0",
    "method": "POST",
    "url": "/api/v1/movies",
    "status": 401,
    "size": 0,
    "duration": 392.254496,
    "time": 1626315682,
    "severity": "INFO",
    "message": "request logged"
}
```

All error logs will have the same request metadata, including `request_id`. The `Request-Id` is also sent back as part of the error response as a response header, allowing you to link the two. An error log will look something like the following:

```json
{
    "level": "error",
    "remote_ip": "127.0.0.1",
    "user_agent": "PostmanRuntime/7.28.0",
    "request_id": "c3nppj6a0brt1dho9e2g",
    "error": "googleapi: Error 401: Request is missing required authentication credential. Expected OAuth 2 access token, login cookie or other valid authentication credential. See https://developers.google.com/identity/sign-in/web/devconsole-project., unauthorized",
    "http_statuscode": 401,
    "realm": "go-api-basic",
    "time": 1626315981,
    "severity": "ERROR",
    "message": "Unauthenticated Request"
}
```

> The above error log demonstrates a log for an error with stack trace turned off.

If the Logger is to be used beyond the scope of the handler, it should be pulled from the request context in the handler and sent as a parameter to any inner calls. The Logger is added only to the request context to capture request related fields with the Logger and be able to pass the initialized logger and middleware handlers easier to the app/route handler. Additional use of the logger should be directly called out in function/method signatures so there are no surprises. All logs from the logger passed down get the benefit of the request metadata though, which is great!

#### Reading and Modifying Logger State

You can retrieve and update the state of these flags using the `{{base_url}}/api/v1/logger` endpoint.

To retrieve the current logger state use a `GET` request:

```bash
curl --location --request GET 'http://127.0.0.1:8080/api/v1/logger' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>'
```

and the response will look something like:

```json
{
    "logger_minimum_level": "debug",
    "global_log_level": "error",
    "log_error_stack": false
}
```

In order to update the logger state use a `PUT` request:

```bash
curl --location --request PUT 'http://127.0.0.1:8080/api/v1/logger' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
--data-raw '{
    "global_log_level": "debug",
    "log_error_stack": "true"
}'
```

and the response will look something like:

```json
{
    "logger_minimum_level": "debug",
    "global_log_level": "debug",
    "log_error_stack": true
}
```

The `PUT` response is the same as the `GET` response, but with updated values. In the examples above, I used a scenario where the logger state started with the global logging level (`global_log_level`) at error and error stack tracing (`log_error_stack`) set to false. The `PUT` request then updates the logger state, setting the global logging level to `debug` and the error stack tracing. You might do something like this if you are debugging an issue and need to see debug logs or error stacks to help with that.

## 7/13/2021 - README under construction

Logging completed. TBD next.
