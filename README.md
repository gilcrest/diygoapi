# DIY Go API

A RESTful API template (built with Go)

The goal of this project is to be an example of a relational database-backed REST HTTP Web Server that has characteristics needed to ensure success in a high volume environment. This project co-opts the DIY ethos of the Go community and does its best to "use the standard library" whenever possible, bringing in third-party libraries when not doing so would be unduly burdensome (structured logging, Oauth2, etc.).

I struggled a lot with parsing the myriad different patterns people have for package layouts and have tried to coalesce what I've learned from others into my own take on a package layout. Below, I hope to communicate how this structure works. If you have any questions, open an issue or send me a note - I'm happy to help! Also, if you disagree or have suggestions, please do the same, I really enjoy getting both positive and negative feedback.

[![Go Reference](https://pkg.go.dev/badge/github.com/gilcrest/diygoapi.svg)](https://pkg.go.dev/github.com/gilcrest/diygoapi) [![Go Report Card](https://goreportcard.com/badge/github.com/gilcrest/diygoapi)](https://goreportcard.com/report/github.com/gilcrest/diygoapi)

## API Walkthrough

The following is an in-depth walkthrough of this project. This is a demo API, so the "business" intent of it is to support basic CRUD (**C**reate, **R**ead, **U**pdate, **D**elete) operations for a movie database. All paths to files or directories are from the project root.

## Minimum Requirements

- [Go](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2/web-server) - authentication
- [Task](https://taskfile.dev/) - task runner for build and script execution
- [CUE](https://cuelang.org/) - config file generation

--------

## Disclaimer

Briefly, the data model for this project is set up to enable a B2B multi-tenant SAAS, which is overkill for a simple CRUD app, however, it's the model I wanted to create and teach myself. That said, it can serve only one tenant just fine.

## Key Terms

- `Person`: from Wikipedia: "A person (plural people or persons) is a being that has certain capacities or attributes such as reason, morality, consciousness or self-consciousness, and being a part of a culturally established form of social relations such as kinship, ownership of property, or legal responsibility. The defining features of personhood and, consequently, what makes a person count as a person, differ widely among cultures and contexts."
- `User`: from Wikipedia: "A user is a person who utilizes a computer or network service." In the context of this project, given that we allow Persons to authenticate with multiple providers, a User is akin to a persona (Wikipedia - "The word persona derives from Latin, where it originally referred to a theatrical mask. On the social web, users develop virtual personas as online identities.") and as such, a Person can have one or many Users (for instance, I can have a GitHub user and a Google user, but I am just one Person). As a general, practical matter, most operations are considered at the User level. For instance, roles are assigned at the user level instead of the Person level, which allows for more fine-grained access control.
- `App`: is an application that interacts with the system. An App always belongs to just one Org.
- `Org`: represents an Organization (company, institution or any other organized body of people with a particular purpose). An Org can have multiple Persons/Users and Apps.

----------

## Getting Started

The following are basic instructions for getting started. 

### Step 1 - Get the code

Clone the code:

```shell
$ git clone https://github.com/gilcrest/diygoapi.git
Cloning into 'diygoapi'...
```

or use the [Github CLI](https://cli.github.com/) (also written in Go!):

```shell
$ gh repo clone gilcrest/diygoapi
Cloning into 'diygoapi'...
```

### Step 2 - Authentication and Authorization

#### Authentication

All requests with this demo webserver require authentication. I have chosen to use [Google's Oauth2 solution](https://developers.google.com/identity/protocols/oauth2/web-server) for these APIs. To use this, you need to setup a Client ID and Client Secret and obtain an access token. The instructions [here](https://developers.google.com/identity/protocols/oauth2) are great.

After Oauth2 setup with Google, I recommend the [Google Oauth2 Playground](https://developers.google.com/oauthplayground/) to obtain fresh access tokens for testing.

Once a user has authenticated through this flow, all calls to services require that the Google access token be sent as a `Bearer` token in the `Authorization` header.

- If there is no token present, an `HTTP 401 (Unauthorized)` response will be sent and the response body will be empty.
- If a token is properly sent, the [Google Oauth2 v2 API](https://pkg.go.dev/google.golang.org/api/oauth2/v2) is used to validate the token. If the token is ***invalid***, an `HTTP 401 (Unauthorized)` response will be sent and the response body will be empty.

> Note: For more details on the authentication model, see the [Authentication Detail](#authentication-detail) section below.

#### Authorization

If the user's Bearer token is ***valid***, the user must be _authorized_. Users must first register with the system and be given a role. Currently, the SelfRegister service accommodates this and automatically assigns a default _role_ `movieAdmin` (functionality will be added eventually for one person registering another).

_Roles_ are assigned _permissions_ and permissions are assigned to resources (service endpoints). The system uses a role-based access control model. The user's role is used to determine if the user is authorized to access a particular endpoint/resource.

The `movieAdmin` role is set up to grant access to all resources. It's a demo... so why not?

> Note: For more details on the authorization model, see the [Authorization Detail](#authorization-detail) section below.

--------

### Step 3 - Configuration

All programs in this project (the web server, database tasks, etc.) use the [ff](https://github.com/peterbourgon/ff) library from [Peter Bourgon](https://peter.bourgon.org) for configuration. The priority order is: **CLI flags > environment variables > config file > defaults**. The config file defaults to `./config/config.json`, so the simplest path is to create that file.

#### Generate a new encryption key

Regardless of which configuration approach you choose, you need a 256-bit ciphertext string, which can be parsed to a 32 byte encryption key. Generate the ciphertext with `task new-key`:

```shell
$ task new-key
Key Ciphertext: [31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06]
```

> Copy the key ciphertext between the brackets to your clipboard to use in one of the options below

#### Option 1 (Recommended for Local Development) - Config File

> Security Disclaimer: Config files make local development easier, however, putting any credentials (encryption keys, username and password, etc.) in a config file is a bad idea from a security perspective. At a minimum, you should have the `config/` directory added to your `.gitignore` file so these configs are not checked in. As this is a template repo, I have checked this all in for example purposes only. The data there is bogus. In an upcoming release, I will integrate with a secrets management platform like [GCP Secret Manager](https://cloud.google.com/secret-manager) or [HashiCorp Vault](https://learn.hashicorp.com/tutorials/vault/getting-started-intro?in=vault/getting-started) [See Issue 91](https://github.com/gilcrest/diygoapi/issues/91).

The config uses a multi-target layout where each target (e.g. `local`, `staging`) has its own settings. Create or edit the JSON file at `./config/config.json`. Update the `encryption_key`, `database` fields (`host`, `port`, `name`, `user`, `password`, `search_path`) and other settings as appropriate for your `PostgreSQL` installation.

```json
{
    "default_target": "local",
    "targets": [
        {
            "target": "local",
            "server_listener_port": 8080,
            "logger": {
                "min_log_level": "trace",
                "log_level": "debug",
                "log_error_stack": false
            },
            "encryption_key": "31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06",
            "database": {
                "host": "localhost",
                "port": 5432,
                "name": "dga_local",
                "user": "demo_user",
                "password": "REPLACE_ME",
                "search_path": "demo"
            }
        }
    ]
}
```

> Setting the [schema search path](https://www.postgresql.org/docs/current/ddl-schemas.html#DDL-SCHEMAS-PATH) properly is critical as the objects in the migration scripts intentionally do not have qualified object names and will therefore use the search path when creating or dropping objects (in the case of the db down migration).

##### Generate config file using CUE (Optional)

If you prefer, you can generate the JSON config file using [CUE](https://cuelang.org/).

The CUE-based config uses a split layout:
- **`config/cue/schema.cue`** -- the shared validation schema (checked into git)
- **`config/config.cue`** -- local config values with credentials (gitignored)
- **`config/config.json`** -- generated output (gitignored)

Edit the `./config/config.cue` file. Update the `encryption_key`, `database` fields (`host`, `port`, `name`, `user`, `password`, `search_path`) and other settings as appropriate for your `PostgreSQL` installation.

After modifying the CUE file, run the following from project root:

```shell
$ task gen-config
```

This should produce the JSON config file mentioned above (at `./config/config.json`).

#### Option 2 - Environment Variables

As an alternative, you can set environment variables directly through bash or whatever strategy you use. Environment variables override config file values. An example bash script:

```bash
#!/bin/bash

# encryption key
export ENCRYPT_KEY="31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06"

# server listen port
export PORT="8080"

# logger environment variables
export LOG_LEVEL_MIN="trace"
export LOG_LEVEL="debug"
export LOG_ERROR_STACK="false"

# Database Environment variables
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="dga_local"
export DB_USER="demo_user"
export DB_PASSWORD="REPLACE_ME"
export DB_SEARCH_PATH="demo"
```

#### Option 3 - Command Line Flags

For full control, you can pass command line flags directly when running a program. Flags take the highest priority, overriding both environment variables and config file values. The following table lists all available flags, their equivalent environment variables, and defaults:

| Flag Name       | Description                                                                                        | Environment Variable | Default   |
|-----------------|----------------------------------------------------------------------------------------------------|----------------------|-----------|
| port            | Port the server will listen on                                                                     | PORT                 | 8080      |
| log-level       | zerolog logging level (debug, info, etc.)                                                          | LOG_LEVEL            | info      |
| log-level-min   | sets the minimum accepted logging level                                                            | LOG_LEVEL_MIN        | trace     |
| log-error-stack | If true, log error stacktrace using github.com/pkg/errors, else just log error (includes op stack) | LOG_ERROR_STACK      | false     |
| db-host         | The host name of the database server.                                                              | DB_HOST              | localhost |
| db-port         | The port number the database server is listening on.                                               | DB_PORT              | 5432      |
| db-name         | The database name.                                                                                 | DB_NAME              |           |
| db-user         | PostgreSQL™ user name to connect as.                                                               | DB_USER              |           |
| db-password     | Password to be used if the server demands password authentication.                                 | DB_PASSWORD          |           |
| db-search-path  | Schema search path to be used when connecting.                                                     | DB_SEARCH_PATH       |           |
| encrypt-key     | Encryption key to be used for all encrypted data.                                                  | ENCRYPT_KEY          |           |

For example:

```bash
$ go run ./cmd/diy/main.go -db-name=dga_local -db-user=demo_user -db-password=REPLACE_ME -db-search-path=demo -encrypt-key=31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06
```

### Step 4 - Database Initialization

The following steps create the database objects and initialize data needed for running the web server. As a convenience, database migration programs which create these objects and load initial data can be executed using [Task](https://taskfile.dev/). To understand database migrations and how they are structured in this project, you can watch [this talk](https://youtu.be/w07butydI5Q) I gave to the [Boston Golang meetup group](https://www.meetup.com/bostongo/?_cookie-check=1Gx8ms5NN8GhlaLJ) in February 2022. The below examples assume you have already setup PostgreSQL and know what user, database and schema you want to install the objects.

> If you want to create an isolated database and schema, you can find examples of doing that at `./scripts/db/db_init.sql`.

> All database tasks read connection info from `./config/config.json` by default, using the `default_target` defined in the config. To target a different environment, pass `--target` via CLI args, e.g. `task db-up -- --target staging`. You can also override the target with the `TARGET` environment variable.

#### Create the Database User

Creating a database user requires elevated privileges. Define a `local-admin` target in your config with a superuser (or a role that has `CREATEROLE`), then run:

```shell
$ TARGET=local-admin task db-create-user
```

This creates the `dga_local` user (with password `REPLACE_ME`) via psql.

#### Run the Database Up Migration

Fifteen database migration scripts are run as part of the up migration:

```shell
$ task db-up
```

> Note: At any time, you can drop all the database objects created as part of the up migration using the down migration scripts in `./scripts/db/migrations/down/`.

#### Data Initialization (Genesis)

There are a number of tables that require initialization of data to facilitate things like: authentication through role based access controls, tracking which applications/users are interacting with the system, etc. I have bundled this initialization into a Genesis service, which can be run only once per database.

TODO - Priority 1! - Talk about calling the Genesis service to setup data

This initial data setup as part of Genesis creates a Principal organization, a Test organization and apps/users within those as well as sets up permissions and roles for access for the user input into the service. The principal org is created solely for the administrative purpose of creating other organizations, apps and users. The test organization is where all tests are run for test data isolation, etc.

Most importantly, a user initiated organization and app is created based on your input. The response details of this organization (located within the `userInitiated` node of the response are those which are needed to run the various Movie APIs (create movie, read movie, etc.)

TODO - show response

--------

### Step 5 - Run Tests

The project tests require that Genesis has been run successfully. If all went well in step 4, you can run the following command to validate:

```shell
$ task test
```

> Note: Some tests require a running database with Genesis data. Packages without database dependencies can be tested independently.

### Step 6 - Run the Web Server

With configuration handled in [Step 3](#step-3---configuration), start the web server with Task:

```shell
$ task run
{"level":"info","time":1675700939,"severity":"INFO","message":"minimum accepted logging level set to trace"}
{"level":"info","time":1675700939,"severity":"INFO","message":"logging level set to debug"}
{"level":"info","time":1675700939,"severity":"INFO","message":"log error stack via github.com/pkg/errors set to false"}
{"level":"info","time":1675700939,"severity":"INFO","message":"sql database opened for localhost on port 5432"}
{"level":"info","time":1675700939,"severity":"INFO","message":"sql database Ping returned successfully"}
{"level":"info","time":1675700939,"severity":"INFO","message":"database version: PostgreSQL 14.6 on aarch64-apple-darwin20.6.0, compiled by Apple clang version 12.0.5 (clang-1205.0.22.9), 64-bit"}
{"level":"info","time":1675700939,"severity":"INFO","message":"current database user: demo_user"}
{"level":"info","time":1675700939,"severity":"INFO","message":"current database: dga_local"}
{"level":"info","time":1675700939,"severity":"INFO","message":"current search_path: demo"}
```

> You can also run directly with `go run ./cmd/diy/main.go`, passing flags or relying on environment variables / config file as described in [Step 3](#step-3---configuration).

### Step 7 - Send Requests

#### cURL Commands to Call Ping Service

With the server up and running, the easiest service to interact with is the `ping` service. This service is a simple health check that returns a series of flags denoting health of the system (queue depths, database up boolean, etc.). For right now, the only thing it checks is if the database is up and pingable. I have left this service unauthenticated so there's at least one service that you can get to without having to have an authentication token, but in actuality, I would typically have every service behind a security token.

Use [cURL](https://curl.se/) GET request to call `ping`:

```bash
$ curl --location --request GET 'http://127.0.0.1:8080/api/v1/ping' \
--header 'x-auth-provider: google' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>'
{"db_up":true}
```

#### cURL Commands to Call Movie Services

The values for the `x-app-id` and `x-api-key` headers needed for all below services are found in the `/api/v1/genesis` service response. The response can be found at `./config/genesis/response.json`:

- APP ID (x-app-id): `userInitiated.app.external_id`
- API Key (x-api-key): `userInitiated.app.api_keys[0].key`

The Bearer token for the `Authorization` header needs to be generated through Google's OAuth2 mechanism. Assuming you've completed setup mentioned in [Step 2](#step-2---authentication-and-authorization), you can generate a new token at the [Google OAuth2 Playground](https://developers.google.com/oauthplayground/)

**Create Movie** - use the `POST` HTTP verb at `/api/v1/movies`:

```shell
$ curl --location --request POST 'http://127.0.0.1:8080/api/v1/movies' \
--header 'Content-Type: application/json' \
--header 'x-app-id: <REPLACE WITH APP ID>' \
--header 'x-api-key: <REPLACE WITH API KEY>' \
--header 'x-auth-provider: google' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
--data-raw '{
    "title": "Repo Man",
    "rated": "R",
    "release_date": "1984-03-02T00:00:00Z",
    "run_time": 92,
    "director": "Alex Cox",
    "writer": "Alex Cox"
}'
{"external_id":"IUAtsOQuLTuQA5OM","title":"Repo Man","rated":"R","release_date":"1984-03-02T00:00:00Z","run_time":92,"director":"Alex Cox","writer":"Alex Cox","create_app_extl_id":"nBRyFTHq6PALwMdx","create_username":"dan@dangillis.dev","create_user_first_name":"Otto","create_user_last_name":"Maddox","create_date_time":"2022-06-30T15:26:02-04:00","update_app_extl_id":"nBRyFTHq6PALwMdx","update_username":"dan@dangillis.dev","update_user_first_name":"Otto","update_user_last_name":"Maddox","update_date_time":"2022-06-30T15:26:02-04:00"}
```

**Read (Single Record)** - use the `GET` HTTP verb at `/api/v1/movies/:extl_id` with the movie `external_id` from the create (POST) response as the unique identifier in the URL. I try to never expose primary keys, so I use something like an external id as an alternative key.

```bash
$ curl --location --request GET 'http://127.0.0.1:8080/api/v1/movies/IUAtsOQuLTuQA5OM' \
--header 'x-app-id: <REPLACE WITH APP ID>' \
--header 'x-api-key: <REPLACE WITH API KEY>' \
--header 'x-auth-provider: google' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
{"external_id":"IUAtsOQuLTuQA5OM","title":"Repo Man","rated":"R","release_date":"1984-03-02T00:00:00Z","run_time":92,"director":"Alex Cox","writer":"Alex Cox","create_app_extl_id":"QfLDvkZlAEieAA7u","create_username":"dan@dangillis.dev","create_user_first_name":"Otto","create_user_last_name":"Maddox","create_date_time":"2022-06-30T15:26:02-04:00","update_app_extl_id":"QfLDvkZlAEieAA7u","update_username":"dan@dangillis.dev","update_user_first_name":"Otto","update_user_last_name":"Maddox","update_date_time":"2022-06-30T15:26:02-04:00"}
```

**Read (All Records)** - use the `GET` HTTP verb at `/api/v1/movies`:

```bash
$ curl --location --request GET 'http://127.0.0.1:8080/api/v1/movies' \
--header 'x-app-id: <REPLACE WITH APP ID>' \
--header 'x-api-key: <REPLACE WITH API KEY>' \
--header 'x-auth-provider: google' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
```

**Update** - use the `PUT` HTTP verb at `/api/v1/movies/:extl_id` with the movie `external_id` from the create (POST) response as the unique identifier in the URL.

```bash
$ curl --location --request PUT 'http://127.0.0.1:8080/api/v1/movies/IUAtsOQuLTuQA5OM' \
--header 'Content-Type: application/json' \
--header 'x-app-id: <REPLACE WITH APP ID>' \
--header 'x-api-key: <REPLACE WITH API KEY>' \
--header 'x-auth-provider: google' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
--data-raw '{
    "title": "Repo Man",
    "rated": "R",
    "release_date": "1984-03-02T00:00:00Z",
    "run_time": 91,
    "director": "Alex Cox",
    "writer": "Alex Cox"
}'
{"external_id":"IUAtsOQuLTuQA5OM","title":"Repo Man","rated":"R","release_date":"1984-03-02T00:00:00Z","run_time":91,"director":"Alex Cox","writer":"Alex Cox","create_app_extl_id":"QfLDvkZlAEieAA7u","create_username":"dan@dangillis.dev","create_user_first_name":"Otto","create_user_last_name":"Maddox","create_date_time":"2022-06-30T15:26:02-04:00","update_app_extl_id":"nBRyFTHq6PALwMdx","update_username":"dan@dangillis.dev","update_user_first_name":"Otto","update_user_last_name":"Maddox","update_date_time":"2022-06-30T15:38:42-04:00"}
```

**Delete** - use the `DELETE` HTTP verb at `/api/v1/movies/:extl_id` with the movie `external_id` from the create (POST) response as the unique identifier in the URL.

```bash
$ curl --location --request DELETE 'http://127.0.0.1:8080/api/v1/movies/IUAtsOQuLTuQA5OM' \
--header 'x-app-id: <REPLACE WITH APP ID>' \
--header 'x-api-key: <REPLACE WITH API KEY>' \
--header 'x-auth-provider: google' \
--header 'Authorization: Bearer <REPLACE WITH ACCESS TOKEN>' \
{"extl_id":"IUAtsOQuLTuQA5OM","deleted":true}
```

--------

## Project Walkthrough

### Package Layout

![RealWorld Example Applications](media/diygoapi-package-layout.png)

The above image is a high-level view of an example request that is processed by the server (creating a movie). To summarize, after receiving an http request, the request path, method, etc. is matched to a registered route in the Server's standard library multiplexer (aka ServeMux, initialization of which, is part of server startup in the `cmd` package as part of the routes.go file in the server package). The request is then sent through a sequence of middleware handlers for setting up request logging, response headers, authentication and authorization. Finally, the request is routed through a bespoke app handler, in this case `handleMovieCreate`.

> `diygoapi` package layout is based on several projects, but the primary source of inspiration is the [WTF Dial app repo](https://github.com/benbjohnson/wtf) and [accompanying blog](https://www.gobeyond.dev/) from [Ben Johnson](https://github.com/benbjohnson). It's really a wonderful resource and I encourage everyone to read it.

### Errors

Handling errors is really important in Go. Errors are first class citizens and there are many different approaches for handling them. I have based my error handling on a [blog post from Rob Pike](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) and have modified it to meet my needs. The post is many years old now, but I find the lessons there still hold true at least for my requirements.

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

All errors should be raised using custom errors from the [errs](https://github.com/gilcrest/diygoapi/tree/main/errs) package. The three custom errors correspond directly to the requirements above.

##### Typical Errors

Typical errors raised throughout `diygoapi` are the custom `errs.Error`, which look like:

 ```go
// Error is the type that implements the error interface.
// It contains a number of fields, each of different type.
// An Error value may leave some values unset.
type Error struct {
    // Op is the operation being performed, usually the name of the method
    // being invoked.
    Op Op
    // User is the name of the user attempting the operation.
    User UserName
    // Kind is the class of error, such as permission failure,
    // or "Other" if its class is unknown or irrelevant.
    Kind Kind
    // Param represents the parameter related to the error.
    Param Parameter
    // Code is a human-readable, short representation of the error
    Code Code
    // Realm is a description of a protected area, used in the WWW-Authenticate header.
    Realm Realm
    // The underlying error that triggered this one, if any.
    Err error
}
```

This custom error type is raised using the `E` function from the [errs](https://github.com/gilcrest/diygoapi/tree/main/errs) package. `errs.E` is taken from Rob Pike's [upspin errors package](https://github.com/upspin/upspin/tree/master/errors) (but has been changed based on my requirements). The `errs.E` function call is [variadic](https://en.wikipedia.org/wiki/Variadic) and can take several different types to form the custom `errs.Error` struct.

Here is a simple example of creating an `error` using `errs.E`:

```go
err := errs.E("seems we have an error here")
```

When a string is sent, a new error will be created and added to the `Err` element of the struct. In the above example, `Op`, `User`, `Kind`, `Param`, `Realm` and `Code` would all remain unset.

By convention, we create an `op` constant to denote the method or function where the error is occuring (or being returned through). This `op` constant should always be the first argument in each call, though it is not actually required to be.

```go
package opdemo

import (
    "fmt"

    "github.com/gilcrest/diygoapi/errs"
)

// IsEven returns an error if the number given is not even
func IsEven(n int) error {
    const op errs.Op = "opdemo/IsEven"

    if n%2 != 0 {
        return errs.E(op, fmt.Sprintf("%d is not even", n))
    }
    return nil
}
```

You can set any of these custom `errs.Error` fields that you like, for example:

```go
var released time.Time
released, err = time.Parse(time.RFC3339, r.Released)
if err != nil {
    return nil, errs.E(op, errs.Validation,
        errs.Code("invalid_date_format"),
        errs.Parameter("release_date"),
        err)
}
```

Above, we used `errs.Validation` to set the `errs.Kind` as `Validation`. Valid error `Kind` are:

```go
const (
    Other          Kind = iota // Unclassified error. This value is not printed in the error message.
    Invalid                    // Invalid operation for this type of item.
    IO                         // External I/O error such as network failure.
    Exist                      // Item already exists.
    NotExist                   // Item does not exist.
    Private                    // Information withheld.
    Internal                   // Internal error or inconsistency.
    BrokenLink                 // Link target does not exist.
    Database                   // Error from database.
    Validation                 // Input validation error.
    Unanticipated              // Unanticipated error.
    InvalidRequest             // Invalid Request
    // Unauthenticated is used when a request lacks valid authentication credentials.
    //
    // For Unauthenticated errors, the response body will be empty.
    // The error is logged and http.StatusUnauthorized (401) is sent.
    Unauthenticated // Unauthenticated Request
    // Unauthorized is used when a user is authenticated, but is not authorized
    // to access the resource.
    //
    // For Unauthorized errors, the response body should be empty.
    // The error is logged and http.StatusForbidden (403) is sent.
    Unauthorized
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
    const op errs.Op = "diygoapi/Movie.IsValid"

    switch {
    case m.Title == "":
        return errs.E(op, errs.Validation, errs.Parameter("title"), errs.MissingField("title"))
```

The error message for the above would read **title is required**

There is also `errs.InputUnwanted` which is meant to be used when a field is populated with a value when it is not supposed to be.

###### Typical Error Flow

As errors created with `errs.E` move up the call stack, the `op` can just be added to the error, like the following:

```go
func outer() error {
    const op errs.Op = "opdemo/outer"

    err := middle()
    if err != nil {
        return errs.E(op, err)
    }
    return nil
}

func middle() error {
    err := inner()
    if err != nil {
        return errs.E(errs.Op("opdemo/middle"), err)
    }
    return nil
}

func inner() error {
    const op errs.Op = "opdemo/inner"

    return errs.E(op, "seems we have an error here")
}
```

> Note that `errs.Op` can be created inline as part of the error instead of creating a constant as done in the middle function, I just prefer to create the constant in most cases.

In addition, you can add context fields (`errs.Code`, `errs.Parameter`, `errs.Kind`) as the error moves up the stack, however, I try to add as much context as possible at the point of error origin and only do this in rare cases.

##### Handler Flow

At the top of the program flow for each route is the handler (for example, [Server.handleMovieCreate](https://github.com/gilcrest/diygoapi/blob/main/server/handlers.go)). In this handler, any error returned from any function or method is sent through the `errs.HTTPErrorResponse` function along with the `http.ResponseWriter` and a `zerolog.Logger`.

For example:

```go
response, err := s.CreateMovieService.Create(r.Context(), rb, u)
if err != nil {
    errs.HTTPErrorResponse(w, logger, err)
    return
}
```

`errs.HTTPErrorResponse` takes the custom `errs.Error` type and writes the response to the given `http.ResponseWriter` and logs the error using the given `zerolog.Logger`.

> `return` must be called immediately after `errs.HTTPErrorResponse` to return the error to the client.

##### Typical Error Response

If an `errs.Error` type is sent to `errs.HTTPErrorResponse`, the function writes the HTTP response body as JSON using the `errs.ErrResponse` struct.

```go
// ErrResponse is used as the Response Body
type ErrResponse struct {
    Error ServiceError `json:"error"`
}

// ServiceError has fields for Service errors. All fields with no data will be omitted
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
        "kind": "input validation error",
        "param": "title",
        "message": "title is required"
    }
}
```

In addition, the error is logged. By default, the error ***stack*** is built using the `op` context added to errors and added to the log as a string array in the `stack` field (see below). For the majority of cases, I believe this is sufficient.

```json
{
   "level": "error",
   "remote_ip": "127.0.0.1:60382",
   "user_agent": "PostmanRuntime/7.30.1",
   "request_id": "cfgihljuns2hhjb77tq0",
   "stack": [
      "diygoapi/Movie.IsValid",
      "service/MovieService.Create"
   ],
   "error": "title is required",
   "http_statuscode": 400,
   "Kind": "input validation error",
   "Parameter": "title",
   "Code": "",
   "time": 1675700438,
   "severity": "ERROR",
   "message": "error response sent to client"
}
```

If you feel you need the full error stack trace, you can set the flag, environment variable on startup or call the `PUT` method for the `{{base_url}}/api/v1/logger` service to update `zerolog.ErrorStackMarshaler` and set it to log error stacks (more about this below). The logger will log the full error stack, which can be super helpful when trying to identify issues.

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

There is logic within `errs.HTTPErrorResponse` to return a different response body if the `errs.Kind` is `Internal` or `Database`. As per the requirements, we should not leak the error message or any internal stack, etc. when an internal or database error occurs. If an error comes through and is an `errs.Error` with either of these error `Kind` or is an unknown error type in any way, the response will look like the following:

```json
{
    "error": {
        "kind": "internal_error",
        "message": "internal server error - please contact support"
    }
}
```

--------

#### Unauthenticated Errors

The [spec](https://tools.ietf.org/html/rfc7235#section-3.1) for `401 Unauthorized` calls for a `WWW-Authenticate` response header along with a `realm`. The realm should be set when creating an Unauthenticated error.

##### Unauthenticated Error Flow

*Unauthenticated* errors should only be raised at points of authentication as part of a middleware handler. I will get into application flow in detail later, but authentication for `diygoapi` happens in middleware handlers prior to calling the final app handler for the given route.

The example below demonstrates returning an *Unauthenticated* error if the Authorization header is not present. This is done using the `errs.E` function (common to all errors in this repo), but the `errs.Kind` is sent as `errs.Unauthenticated`. An `errs.Realm` type should be added as well. For now, the constant `defaultRealm` is set to `diygoapi` in the `server` package and is used for all unauthenticated errors. You can set this constant to whatever value you like for your application.

```go
// parseAuthorizationHeader parses/validates the Authorization header and returns an Oauth2 token
func parseAuthorizationHeader(realm string, header http.Header) (*oauth2.Token, error) {
    const op errs.Op = "server/parseAuthorizationHeader"

    // Pull the token from the Authorization header by retrieving the
    // value from the Header map with "Authorization" as the key
    //
    // format: Authorization: Bearer
    headerValue, ok := header["Authorization"]
    if !ok {
        return nil, errs.E(op, errs.Unauthenticated, errs.Realm(realm), "unauthenticated: no Authorization header sent")
    }
...
```

##### Unauthenticated Error Response

Per requirements, `diygoapi` does not return a response body when returning an **Unauthenticated** error. The error response from [cURL](https://curl.se/) looks like the following:

```bash
HTTP/1.1 401 Unauthorized
Request-Id: c30hkvua0brkj8qhk3e0
Www-Authenticate: Bearer realm="diygoapi"
Date: Wed, 09 Jun 2021 19:46:07 GMT
Content-Length: 0
```

--------

#### Unauthorized Errors

If the user is not authorized to use the API, an `HTTP 403 (Forbidden)` response will be sent and the response body will be empty.

##### Unauthorized Error Flow

*Unauthorized* errors are raised when there is a permission issue for a user attempting to access a resource. `diygoapi` currently has a custom database-driven RBAC (Role Based Access Control) authorization implementation (more about this later). The below example demonstrates raising an *Unauthorized* error and is found in the `DBAuthorizer.Authorize` method.

```go
return errs.E(errs.Unauthorized, fmt.Sprintf("user %s does not have %s permission for %s", adt.User.Username, r.Method, pathTemplate))
```

Per requirements, `diygoapi` does not return a response body when returning an **Unauthorized** error. The error response from [cURL](https://curl.se/) looks like the following:

```bash
HTTP/1.1 403 Forbidden
Request-Id: c30hp2ma0brkj8qhk3f0
Date: Wed, 09 Jun 2021 19:54:50 GMT
Content-Length: 0
```

### Logging

`diygoapi` uses the [zerolog](https://github.com/rs/zerolog) library from [Olivier Poitrey](https://github.com/rs). The mechanics for using `zerolog` are straightforward and are well documented in the library's [README](https://github.com/rs/zerolog#readme). `zerolog` takes an `io.Writer` as input to create a new logger; for simplicity in `diygoapi`, I use `os.Stdout`.

#### Setting Logger State on Startup

When starting `diygoapi`, there are several flags which setup the logger:

| Flag Name       | Description                                                                                        | Environment Variable | Default |
|-----------------|----------------------------------------------------------------------------------------------------|----------------------|---------|
| log-level       | zerolog logging level (debug, info, etc.)                                                          | LOG_LEVEL            | info    |
| log-level-min   | sets the minimum accepted logging level                                                            | LOG_LEVEL_MIN        | trace   |
| log-error-stack | If true, log error stacktrace using github.com/pkg/errors, else just log error (includes op stack) | LOG_ERROR_STACK      | false   |

--------

> As mentioned in [Step 3](#step-3---configuration), `diygoapi` uses the [ff](https://github.com/peterbourgon/ff) library from [Peter Bourgon](https://peter.bourgon.org), which allows for using flags, environment variables, or a config file. Going forward, we'll assume you've chosen flags.

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

The `log-error-stack` boolean flag tells whether to log full stack traces for each error. If `true`, the `zerolog.ErrorStackMarshaler` will be set to `pkgerrors.MarshalStack` which means, for errors raised using the [github.com/pkg/errors](https://github.com/pkg/errors) package, the error stack trace will be captured and printed along with the log. All errors raised in `diygoapi` are raised using `github.com/pkg/errors` if this flag is set to true.

After parsing the command line flags, `zerolog.Logger` is initialized in `cmd/cmd.go`

```go
// setup logger with appropriate defaults
lgr := logger.NewWithGCPHook(os.Stdout, minlvl, true)
```

and subsequently used to initialize the `server.Server` struct.

```go
// initialize Server enfolding a http.Server with default timeouts,
// a mux router and a zerolog.Logger
s := server.New(http.NewServeMux(), server.NewDriver(), lgr)
```

#### Logger Setup in Handlers

The `Server.registerRoutes` method is responsible for registering routes and corresponding middleware/handlers to the Server's multiplexer (aka router). For each route registered to the handler, upon execution, the initialized `zerolog.Logger` struct is added to the request context through the `Server.loggerChain` method.

```go
// register routes/middleware/handlers to the Server ServeMux
func (s *Server) registerRoutes() {

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	s.mux.Handle("POST /api/v1/movies",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleMovieCreate))

...
```

The `Server.loggerChain` method sets up the logger with pre-populated fields, including the request method, url, status, size, duration, remote IP, user agent, referer. A unique `Request ID` is also added to the logger, context and response headers.

```go
func (s *Server) loggerChain() alice.Chain {
	ac := alice.New(hlog.NewHandler(s.Logger),
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
   "remote_ip": "127.0.0.1:60382",
   "user_agent": "PostmanRuntime/7.30.1",
   "request_id": "cfgihljuns2hhjb77tq0",
   "method": "POST",
   "url": "/api/v1/movies",
   "status": 400,
   "size": 90,
   "duration": 85.747943,
   "time": 1675700438,
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
    "realm": "diygoapi",
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

#### Authentication Detail

**Authentication** is determined by validating the `App` making the request as well as the `User`.

##### App Authentication Detail

An `App` (aka API Client) can be registered using the `POST /api/v1/apps` service. The `App` can be registered with an association to an Oauth2 Provider Client ID or be standalone.

An `App` has two possible methods of authentication.

1. The first method, which overrides the second, is using the `X-APP-ID` and `X-API-KEY` HTTP headers. The `X-APP-ID` is the app unique identifier and the `X-API-KEY` is the password. This method confirms the veracity of the App against values stored in the database (the password is encrypted in the db). If the App ID is not found or the API key does not match the stored API key, an `HTTP 401 (Unauthorized)` response will be sent and the response body will be empty. If the authentication is successful, the App details will be set to the request context for downstream use.
2. If there is no X-APP-ID header present, the second method is using the authorization Provider's Client ID associated with the `Authorization` header Bearer token. When a request is made only using the `Authorization` header, a callback to the provider's Oauth2 TokenInfo API is done to retrieve the associated Provider Client ID. The Provider Client ID is then used to find the associated App in the database. If the App is not found, an `HTTP 401 (Unauthorized)` response will be sent and the response body will be empty. If the App is found, the App details will be set to the request context for downstream use.

##### User Authentication Detail

User authentication happens outside the API using an Oauth2 provider (e.g. Google, Github, etc.) sign-in page. After successful authentication, the user is given a Bearer token which is then used for service-level authentication.

In order to perform any actions with the `diygoapi` services, a User must Self-Register. The `SelfRegister` service creates a Person/User and stores them in the database. In addition, an `Auth` object, which represents the user's credentials is stored in the db. A search is done prior to `Auth` creation to determine if the user is already registered, and if so, the existing user is returned.

After a user is registered, they can perform actions (use resources/endpoints) using the services. For every call, the `Authorization` HTTP header with the user's Bearer token along with the `X-AUTH-PROVIDER` header used to denote the `Provider` (e.g. Google, Github, etc.) must be set.

The Bearer token is used to find the `Auth` object for the user in the database. Searching for the `Auth` object is done as follows:
- Search the database directly using the Bearer token.
  - If an `Auth` object already exists in the datastore which matches the `Bearer` token and the `Bearer` token is not past its expiration date, the existing `Auth` will be used to determine the User.
  - If no `Auth` object exists in the datastore for the request `Bearer` token, an attempt will be made to find the user's `Auth` with the Provider ID (e.g. Google) and unique person ID given by the provider (found by calling the provider API with the request `Bearer` token). If an `Auth` object exists given these attributes, it will be updated with the new `Bearer` token details and this `Auth` will be used to obtain the `User` details.

If the `Auth` object is not found, an `HTTP 401 (Unauthorized)` response will be sent and the response body will be empty. If the `Auth` object is found, the `User` details will be set to the request context for downstream use.

#### Authorization Detail

TODO

##### Role Based Access Control

TODO