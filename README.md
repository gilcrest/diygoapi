# DIY Go API

A RESTful API template (built with Go)

The goal of this project is to be an example of a relational database-backed REST HTTP Web Server that has characteristics needed to ensure success in a high volume environment. This project co-opts the DIY ethos of the Go community and does its best to "use the standard library" whenever possible, bringing in third-party libraries when not doing so would be unduly burdensome (structured logging, Oauth2, etc.).

I struggled a lot with parsing the myriad different patterns people have for package layouts over the past few years and have tried to coalesce what I've learned from others into my own take on a package layout. Below, I hope to communicate how this structure works. If you have any questions, open an issue or send me a note - I'm happy to help! Also, if you disagree or have suggestions, please do the same, I really enjoy getting both positive and negative feedback.

[![Go Reference](https://pkg.go.dev/badge/github.com/gilcrest/diygoapi.svg)](https://pkg.go.dev/github.com/gilcrest/diygoapi) [![Go Report Card](https://goreportcard.com/badge/github.com/gilcrest/diygoapi)](https://goreportcard.com/report/github.com/gilcrest/diygoapi)

## API Walkthrough

The following is an in-depth walkthrough of this project. This is a demo API, so the "business" intent of it is to support basic CRUD (**C**reate, **R**ead, **U**pdate, **D**elete) operations for a movie database. All paths to files or directories are from the project root.

## Minimum Requirements

- [Go](https://go.dev/)
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2/web-server) - authentication
- [Mage](https://magefile.org/) - for build and easier script execution
- [CUE](https://cuelang.org/) - config file generation

--------

## Getting Started

The following are basic instructions for getting started. For detailed explanations of many of the constructs created as part of these steps, jump to the [Project Walkthrough](#project-walkthrough)

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

#### Authorization

- If the token is ***valid***, Google will respond with information about the user. The user's email will be used as their username in addition to determining if the user is authorized for access to a particular endpoint/resource. The authorization is done through an internal database role based access control model. If the user is not authorized to use the API, an `HTTP 403 (Forbidden)` response will be sent and the response body will be empty.

--------

### Step 3 - Prepare Environment (2 options)

All Mage programs in this project which take an environment (env) parameter (e.g., `func DBUp(env string)`), must have certain environment variables set. These environment variables can be set independently [option 1](#option-1---set-your-environment-independently) or based on a configuration file [option 2](#option-2---set-your-environment-through-a-config-file). Depending on which environment method you choose, the values to pass to the env parameter when running Mage programs in this project are as follows:

| env string | File Path             | Description                                                                                                    |
|------------|-----------------------|----------------------------------------------------------------------------------------------------------------|
| current    | N/A                   | Uses the current session environment. Environment will not be overriden from a config file                     |
| local      | ./config/local.json   | Uses the `local.json` config file to set the environment                                                       |
| staging    | ./config/staging.json | Uses the `staging.json` config file to set the environment in [Google Cloud Run](https://cloud.google.com/run) |

The base environment variables to be set are:

| Environment Variable | Description                                                        |
|----------------------|--------------------------------------------------------------------|
| PORT                 | Port the server will listen on                                     |
| LOG_LEVEL            | zerolog logging level (debug, info, etc.)                          |
| LOG_LEVEL_MIN        | sets the minimum accepted logging level                            |
| LOG_ERROR_STACK      | If true, log error stacktrace using github.com/pkg/errors, else just log error (includes op stack) |
| DB_HOST              | The host name of the database server.                              |
| DB_PORT              | The port number the database server is listening on.               |
| DB_NAME              | The database name.                                                 |
| DB_USER              | PostgreSQL™ user name to connect as.                               |
| DB_PASSWORD          | Password to be used if the server demands password authentication. |
| DB_SEARCH_PATH       | Schema Search Path                                                 |
| ENCRYPT_KEY          | Encryption Key                                                     |

> The same environment variables are used when running the web server, but are not mandatory. When running the web server, if you prefer, you can bypass environment variables and instead send command line flags (more about that later).

#### Generate a new encryption key

Either option below for setting the environment requires a 256-bit ciphertext string, which can be parsed to a 32 byte encryption key. Generate the ciphertext with the `NewKey` mage program:

```shell
$ mage -v newkey
Running target: NewKey
Key Ciphertext: [31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06]
```

> Copy the key ciphertext between the brackets to your clipboard to use in option 1 or 2 below

#### Option 1 - Set your environment independently

As always, you can set your environment on your own through bash or whatever strategy you use for this, an example bash script should you choose:

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

#### Option 2 - Set your environment through a Config File

##### Generate new config file using CUE

Another option is to use a JSON configuration file generated by [CUE](https://cuelang.org/) located at `./config/local.json`.

In order to generate this file, edit the `./config/cue/local.cue` file. Paste and overwrite the ciphertext from your clipboard into the `config: encryptionKey:` field of the file and also update the `config: database:` fields (`host`, `port`, `name`, `user`, `password`, `searchPath`) as appropriate for your `PostgreSQL` installation.

```cue
package config

config: #LocalConfig

config: encryptionKey: "31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06"

config: httpServer: listenPort: 8080

config: logger: minLogLevel:   "trace"
config: logger: logLevel:      "debug"
config: logger: logErrorStack: false

config: database: host:       "localhost"
config: database: port:       5432
config: database: name:       "dga_local"
config: database: user:       "demo_user"
config: database: password:   "REPLACE_ME"
config: database: searchPath: "demo"
```

> Security Disclaimer: Config files make local development easier, however, putting any credentials (encryption keys, username and password, etc.) in a config file is a bad idea from a security perspective. At a minimum, you should have the `config/` directory added to your `.gitignore` file so these configs are not checked in. As this is a template repo, I have checked this all in for example purposes only. The data there is bogus. In an upcoming release, I will integrate with a secrets management platform like [GCP Secret Manager](https://cloud.google.com/secret-manager) or [HashiCorp Vault](https://learn.hashicorp.com/tutorials/vault/getting-started-intro?in=vault/getting-started) [Issue 91](https://github.com/gilcrest/diygoapi/issues/91).

After modifying the above file, run the following from project root:

```shell
$ mage -v cueGenerateConfig local
Running target: CueGenerateConfig
exec: cue "vet" "./config/cue/schema.cue" "./config/cue/local.cue"
exec: cue "fmt" "./config/cue/schema.cue" "./config/cue/local.cue"
exec: cue "export" "./config/cue/schema.cue" "./config/cue/local.cue" "--force" "--out" "json" "--outfile" "./config/local.json"
```

`cueGenerateConfig` should produce a JSON config file at `./config/local.json` that looks similar to:

```json
{
    "config": {
        "httpServer": {
            "listenPort": 8080
        },
        "logger": {
            "minLogLevel": "trace",
            "logLevel": "debug",
            "logErrorStack": false
        },
        "encryptionKey": "31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06",
        "database": {
            "host": "localhost",
            "port": 5432,
            "name": "dga_local",
            "user": "demo_user",
            "password": "REPLACE_ME",
            "searchPath": "demo"
        }
    }
}
```

> Setting the [schema search path](https://www.postgresql.org/docs/current/ddl-schemas.html#DDL-SCHEMAS-PATH) properly is critical as the objects in the migration scripts intentionally do not have qualified object names and will therefore use the search path when creating or dropping objects (in the case of the db down migration).

### Step 4 - Database Initialization

The following steps setup the database objects and initialize data needed for running the web server. As a convenience, database migration programs which create these objects and load initial data can be executed using [Mage](https://magefile.org/). To understand database migrations and how they are structured in this project, you can watch [this talk](https://youtu.be/w07butydI5Q) I gave to the [Boston Golang meetup group](https://www.meetup.com/bostongo/?_cookie-check=1Gx8ms5NN8GhlaLJ) in February 2022. The below examples assume you have already setup PostgreSQL and know what user, database and schema you want to install the objects.

> If you want to create an isolated database and schema, you can find examples of doing that at `./scripts/db/db_init.sql`.

#### Run the Database Up Migration

Twelve database tables are created as part of the up migration.

```shell
$ mage -v dbup local
Running target: DBUp
exec: psql "-w" "-d" "postgresql://demo_user@localhost:5432/dga_local?options=-csearch_path%3Ddemo" "-c" "select current_database(), current_user, version()" "-f" "./scripts/db/migrations/up/001-app.sql" "-f" "./scripts/db/migrations/up/002-org_user.sql" "-f" "./scripts/db/migrations/up/003-permission.sql" "-f" "./scripts/db/migrations/up/004-person.sql" "-f" "./scripts/db/migrations/up/005-org_kind.sql" "-f" "./scripts/db/migrations/up/006-role.sql" "-f" "./scripts/db/migrations/up/014-movie.sql" "-f" "./scripts/db/migrations/up/008-app_api_key.sql" "-f" "./scripts/db/migrations/up/009-person_profile.sql" "-f" "./scripts/db/migrations/up/010-org.sql" "-f" "./scripts/db/migrations/up/011-role_permission.sql" "-f" "./scripts/db/migrations/up/012-role_user.sql"
 current_database | current_user |                                                      version                                                      
------------------+--------------+-------------------------------------------------------------------------------------------------------------------
 dga_local        | demo_user    | PostgreSQL 14.2 on aarch64-apple-darwin20.6.0, compiled by Apple clang version 12.0.5 (clang-1205.0.22.9), 64-bit
(1 row)

CREATE TABLE
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
COMMENT
CREATE INDEX
CREATE INDEX
CREATE TABLE
COMMENT
COMMENT
COMMENT
...
```

> Note: At any time, you can drop all the database objects created as part of the up migration using using the down migration program: `mage -v dbdown local`

#### Data Initialization (Genesis)

There are a number of tables that require initialization of data to facilitate things like: authentication through role based access controls, tracking which applications/users are interacting with the system, etc. I have bundled this initialization into a Genesis service, which can be run only once per database. This can be run as a service, but for ease of use, there is a mage program for it as well.

The `genesis` mage program uses a JSON configuration file generated by [CUE](https://cuelang.org/) located at `./config/genesis/request.json`.

To generate this file, navigate to `./config/genesis/cue/input.cue` and update the user details you plan to authenticate with via Google Oauth2. If you wish, you can update the initial org and app details from the default values you'll find in the file as well:

```cue
package genesis

// The "genesis" user - the first user to create the system and is
// given the sysAdmin role (which has all permissions). This user is
// added to the Principal org and the user initiated org created below.
user: email:      "otto.maddox@gmail.com"
user: first_name: "Otto"
user: last_name:  "Maddox"

// The first organization created which can actually transact
// (e.g. is not the principal or test org)
org: name:        "Movie Makers Unlimited"
org: description: "An organization dedicated to creating movies in a demo app."
org: kind:        "standard"

// The initial app created along with the Organization created above
org: app: name:        "Movie Makers App"
org: app: description: "The first app dedicated to creating movies in a demo app."
```

Next, use mage to run the `cueGenerateGenesisConfig` program:

```shell
$ mage -v cueGenerateGenesisConfig
Running target: CueGenerateGenesisConfig
exec: cue "vet" "./config/genesis/cue/schema.cue" "./config/genesis/cue/auth.cue" "./config/genesis/cue/input.cue"
exec: cue "fmt" "./config/genesis/cue/schema.cue" "./config/genesis/cue/auth.cue" "./config/genesis/cue/input.cue"
exec: cue "export" "./config/genesis/cue/schema.cue" "./config/genesis/cue/auth.cue" "./config/genesis/cue/input.cue" "--force" "--out" "json" "--outfile" "./config/genesis/request.json"
```

This will generate `./config/genesis/request.json` similar to the below. This file also includes information about which permissions and roles to create as part of Genesis. Leave those as is.

```json
{
    "user": {
        "email": "otto.maddox@gmail.com",
        "first_name": "Otto",
        "last_name": "Maddox"
    },
    "org": {
        "name": "Movie Makers Unlimited",
        "description": "An organization dedicated to creating movies in a demo app.",
        "kind": "standard",
        "app": {
            "name": "Movie Makers App",
            "description": "The first app dedicated to creating movies in a demo app."
        }
    },
    "permissions": [
        {
            "resource": "/api/v1/ping",
            "operation": "GET",
            "description": "allows for calling the ping service to determine if system is up and running",
            "active": true
        },
        {
            "resource": "/api/v1/logger",
            "operation": "GET",
            "description": "allows for reading the logger state",
            "active": true
        },
        {
            "resource": "/api/v1/logger",
            "operation": "PUT",
            "description": "allows for updating the logger state",
            "active": true
        },
        {
            "resource": "/api/v1/orgs",
            "operation": "POST",
...
```

Execute the `Genesis` mage program to initialize the database with dependent data:

```shell
$ mage -v genesis local
Running target: Genesis
{"level":"info","time":1654723891,"severity":"INFO","message":"minimum accepted logging level set to trace"}
{"level":"info","time":1654723891,"severity":"INFO","message":"logging level set to debug"}
{"level":"info","time":1654723891,"severity":"INFO","message":"log error stack global set to true"}
{"level":"info","time":1654723891,"severity":"INFO","message":"sql database opened for localhost on port 5432"}
{"level":"info","time":1654723891,"severity":"INFO","message":"sql database Ping returned successfully"}
{"level":"info","time":1654723891,"severity":"INFO","message":"database version: PostgreSQL 14.2 on aarch64-apple-darwin20.6.0, compiled by Apple clang version 12.0.5 (clang-1205.0.22.9), 64-bit"}
{"level":"info","time":1654723891,"severity":"INFO","message":"current database user: demo_user"}
{"level":"info","time":1654723891,"severity":"INFO","message":"current database: dga_local"}
{"level":"info","time":1654723891,"severity":"INFO","message":"current search_path: demo"}
{
  "principal": {
    "org": {
      "external_id": "HmiB9CmMpUU8hdVk",
      "name": "Principal",
      "kind_description": "genesis",
      "description": "The Principal org represents the first organization created in the database and exists for the administrative purpose of creating other organizations, apps and users.",
      "create_app_extl_id": "L-qGp1UquEgxKjn2",
      "create_username": "otto.maddox@gmail.com",
      "create_user_first_name": "Otto",
      "create_user_last_name": "Maddox",
      "create_date_time": "2022-06-08T17:31:31-04:00",
      "update_app_extl_id": "L-qGp1UquEgxKjn2",
      "update_username": "otto.maddox@gmail.com",
      "update_user_first_name": "Otto",
      "update_user_last_name": "Maddox",
      "update_date_time": "2022-06-08T17:31:31-04:00"
    },
    "app": {
      "external_id": "L-qGp1UquEgxKjn2",
      "name": "Developer Dashboard",
      "description": "App created as part of Genesis event. To be used solely for creating other apps, orgs and users.",
      "create_app_extl_id": "L-qGp1UquEgxKjn2",
      "create_username": "otto.maddox@gmail.com",
      "create_user_first_name": "Otto",
      "create_user_last_name": "Maddox",
      "create_date_time": "2022-06-08T17:31:31-04:00",
      "update_app_extl_id": "L-qGp1UquEgxKjn2",
      "update_username": "otto.maddox@gmail.com",
      "update_user_first_name": "Otto",
      "update_user_last_name": "Maddox",
      "update_date_time": "2022-06-08T17:31:31-04:00",
      "api_keys": [
        {
          "key": "ZXo3BL-deFqP2VXLIYDAbZzF",
          "deactivation_date": "2099-12-31 00:00:00 +0000 UTC"
        }
      ]
    }
  },
  "test": {
    "org": {
...
```

When running the `Genesis` service through `mage`, the JSON response is sent to the terminal and also `./config/genesis/response.json` so you don't need to collect it now.

Briefly, the data model is setup to enable a B2B multi-tenant SAAS, which is overkill for a simple CRUD app, but it's the model I wanted to create/learn and can serve only one tenant just fine. This initial data setup as part of Genesis creates a Principal organization, a Test organization and apps/users within those as well as sets up permissions and roles for access for the user input into the service. The principal org is created solely for the administrative purpose of creating other organizations, apps and users. The test organization is where all tests are run for test data isolation, etc.

Most importantly, a user initiated organization and app is created based on your input in `./config/genesis/cue/input.cue`. The response details of this organization (located within the `userInitiated` node of the response are those which are needed to run the various Movie APIs (create movie, read movie, etc.)

--------

### Step 5 - Run Tests

The project tests require that Genesis has been run successfully. If all went well in step 4, you can run the following command to validate:

```shell
$ mage -v testall false local
Running target: TestAll
exec: go "test" "./..."
?       github.com/gilcrest/diygoapi  [no test files]
?       github.com/gilcrest/diygoapi/app      [no test files]
?       github.com/gilcrest/diygoapi/audit    [no test files]
ok      github.com/gilcrest/diygoapi/auth     0.331s
ok      github.com/gilcrest/diygoapi/command  0.724s
ok      github.com/gilcrest/diygoapi/datastore        0.682s
?       github.com/gilcrest/diygoapi/datastore/appstore       [no test files]
?       github.com/gilcrest/diygoapi/datastore/authstore      [no test files]
?       github.com/gilcrest/diygoapi/datastore/datastoretest  [no test files]
?       github.com/gilcrest/diygoapi/datastore/moviestore     [no test files]
?       github.com/gilcrest/diygoapi/datastore/orgstore       [no test files]
?       github.com/gilcrest/diygoapi/datastore/personstore    [no test files]
?       github.com/gilcrest/diygoapi/datastore/pingstore      [no test files]
ok      github.com/gilcrest/diygoapi/datastore/userstore      0.742s
ok      github.com/gilcrest/diygoapi/errs     0.490s
?       github.com/gilcrest/diygoapi/gateway  [no test files]
?       github.com/gilcrest/diygoapi/gateway/authgateway      [no test files]
ok      github.com/gilcrest/diygoapi/logger   0.284s
?       github.com/gilcrest/diygoapi/magefiles        [no test files]
ok      github.com/gilcrest/diygoapi/movie    0.571s
?       github.com/gilcrest/diygoapi/org      [no test files]
?       github.com/gilcrest/diygoapi/person   [no test files]
ok      github.com/gilcrest/diygoapi/random   0.333s
?       github.com/gilcrest/diygoapi/random/randomtest        [no test files]
ok      github.com/gilcrest/diygoapi/secure   0.574s
?       github.com/gilcrest/diygoapi/secure/random    [no test files]
ok      github.com/gilcrest/diygoapi/server   0.328s
?       github.com/gilcrest/diygoapi/server/driver    [no test files]
ok      github.com/gilcrest/diygoapi/service  0.504s
ok      github.com/gilcrest/diygoapi/user     0.323s
?       github.com/gilcrest/diygoapi/user/usertest    [no test files]
```

> Note: There are a number of packages without test files, but there is extensive testing as part of this project. More can and will be done, of course...

### Step 6 - Run the Web Server

There are three options for running the web server. When running the program, a number of flags can be passed instead of using the environment. The [ff](https://github.com/peterbourgon/ff) library from [Peter Bourgon](https://peter.bourgon.org) is used to parse the flags. If your preference is to set configuration with [environment variables](https://en.wikipedia.org/wiki/Environment_variable), that is possible as well. Flags take precedence, so if a flag is passed, that will be used. A PostgreSQL database connection is required. If there is no flag set, then the program checks for a matching environment variable. If neither are found, the flag's default value will be used and, depending on the flag, may result in a database connection error.

For simplicity’s sake, the easiest option to start with is setting the environment and running the server with Mage:

#### Option 1 - Run web server with config file and Mage

You can run the webserver with Mage. As in all examples above, Mage will either use the current environment or set the environment using a config file depending on the environment parameter sent in.

```shell
$ mage -v run local
Running target: Run
exec: go "run" "./cmd/diy/main.go"
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

#### Option 2 - Run web server using command line flags

The below are the list of the command line flags that can be used to start the webserver (and their equivalent environment variable name for reference as well):

| Flag Name       | Description                                                        | Environment Variable | Default |
|-----------------|--------------------------------------------------------------------|----------------------|---------|
| port            | Port the server will listen on                                     | PORT                 | 8080    |
| log-level       | zerolog logging level (debug, info, etc.)                          | LOG_LEVEL            | debug   |
| log-level-min   | sets the minimum accepted logging level                            | LOG_LEVEL_MIN        | debug   |
| log-error-stack | If true, log error stacktrace using github.com/pkg/errors, else just log error (includes op stack) | LOG_ERROR_STACK      | false   |
| db-host         | The host name of the database server.                              | DB_HOST              |         |
| db-port         | The port number the database server is listening on.               | DB_PORT              | 5432    |
| db-name         | The database name.                                                 | DB_NAME              |         |
| db-user         | PostgreSQL™ user name to connect as.                               | DB_USER              |         |
| db-password     | Password to be used if the server demands password authentication. | DB_PASSWORD          |         |
| db-search-path  | Schema search path to be used when connecting.                     | DB_SEARCH_PATH       |         |
| encrypt-key     | Encryption key to be used for all encrypted data.                  | ENCRYPT_KEY          |         |

Starting the web server with command line flags looks like:

```bash
$ go run main.go -db-name=dga_local -db-user=demo_user -db-password=REPLACE_ME -db-search-path=demo -encrypt-key=31f8cbffe80df0067fbfac4abf0bb76c51d44cb82d2556743e6bf1a5e25d4e06
{"level":"info","time":1656296193,"severity":"INFO","message":"minimum accepted logging level set to trace"}
{"level":"info","time":1656296193,"severity":"INFO","message":"logging level set to debug"}
{"level":"info","time":1656296193,"severity":"INFO","message":"log error stack global set to true"}
{"level":"info","time":1656296193,"severity":"INFO","message":"sql database opened for localhost on port 5432"}
{"level":"info","time":1656296193,"severity":"INFO","message":"sql database Ping returned successfully"}
{"level":"info","time":1656296193,"severity":"INFO","message":"database version: PostgreSQL 14.4 on aarch64-apple-darwin20.6.0, compiled by Apple clang version 12.0.5 (clang-1205.0.22.9), 64-bit"}
{"level":"info","time":1656296193,"severity":"INFO","message":"current database user: demo_user"}
{"level":"info","time":1656296193,"severity":"INFO","message":"current database: dga_local"}
{"level":"info","time":1656296193,"severity":"INFO","message":"current search_path: demo"}
```

#### Option 3 - Run web server using independently set environment

If you're not using mage or command line flags and have set the appropriate environment variables properly, you can run the web server simply like so:

```bash
$ go run main.go
{"level":"info","time":1656296765,"severity":"INFO","message":"minimum accepted logging level set to trace"}
{"level":"info","time":1656296765,"severity":"INFO","message":"logging level set to debug"}
{"level":"info","time":1656296765,"severity":"INFO","message":"log error stack global set to true"}
{"level":"info","time":1656296765,"severity":"INFO","message":"sql database opened for localhost on port 5432"}
{"level":"info","time":1656296765,"severity":"INFO","message":"sql database Ping returned successfully"}
{"level":"info","time":1656296765,"severity":"INFO","message":"database version: PostgreSQL 14.4 on aarch64-apple-darwin20.6.0, compiled by Apple clang version 12.0.5 (clang-1205.0.22.9), 64-bit"}
{"level":"info","time":1656296765,"severity":"INFO","message":"current database user: gilcrest"}
{"level":"info","time":1656296765,"severity":"INFO","message":"current database: dga_local"}
{"level":"info","time":1656296765,"severity":"INFO","message":"current search_path: demo"}
```

### Step 7 - Send Requests

#### cURL Commands to Call Ping Service

With the server up and running, the easiest service to interact with is the `ping` service. This service is a simple health check that returns a series of flags denoting health of the system (queue depths, database up boolean, etc.). For right now, the only thing it checks is if the database is up and pingable. I have left this service unauthenticated so there's at least one service that you can get to without having to have an authentication token, but in actuality, I would typically have every service behind a security token.

Use [cURL](https://curl.se/) GET request to call `ping`:

```bash
$ curl --location --request GET 'http://127.0.0.1:8080/api/v1/ping'
{"db_up":true}
```

#### cURL Commands to Call Movie Services

The values for the `x-app-id` and `x-api-key` headers needed for all below services are found in the `/api/v1/genesis` service response. If you used `mage` to run the service on your local machine, the response can be found at `./config/genesis/response.json`:

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

The above image is a high-level view of an example request that is processed by the server (creating a movie). To summarize, after receiving an http request, the request path, method, etc. is matched to a registered route in the [gorilla mux](https://github.com/gorilla/mux) router (router initialization is part of server startup in the `command` package) as part of the routes.go file in the server package. The request is then sent through a sequence of middleware handlers for setting up request logging, response headers, authentication and authorization. Finally, the request is routed through a bespoke app handler, in this case `handleMovieCreate`.

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

All errors should be raised using custom errors from the [domain/errs](https://github.com/gilcrest/diygoapi/tree/main/domain/errs) package. The three custom errors correspond directly to the requirements above.

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

This custom error type is raised using the `E` function from the [domain/errs](https://github.com/gilcrest/diygoapi/tree/main/domain/errs) package. `errs.E` is taken from Rob Pike's [upspin errors package](https://github.com/upspin/upspin/tree/master/errors) (but has been changed based on my requirements). The `errs.E` function call is [variadic](https://en.wikipedia.org/wiki/Variadic) and can take several different types to form the custom `errs.Error` struct.

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

*Unauthorized* errors are raised when there is a permission issue for a user attempting to access a resource. `diygoapi` currently has a custom database-driven RBAC (Role Based Access Control) authorization implementation (more about this later). The below example demonstrates raising an *Unauthorized* error and is found in the [DBAuthorizer.Authorize](https://github.com/gilcrest/diygoapi/blob/v0.47.3/service/rbac.go#L37) method.

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

| Flag Name       | Description                                             | Environment Variable | Default |
|-----------------|---------------------------------------------------------|----------------------|---------|
| log-level       | zerolog logging level (debug, info, etc.)               | LOG_LEVEL            | debug   |
| log-level-min   | sets the minimum accepted logging level                 | LOG_LEVEL_MIN        | debug   |
| log-error-stack | If true, log error stacktrace using github.com/pkg/errors, else just log error (includes op stack) | LOG_ERROR_STACK      | false   |

--------

> As mentioned [above](https://github.com/gilcrest/diygoapi#command-line-flags), `diygoapi` uses the [ff](https://github.com/peterbourgon/ff) library from [Peter Bourgon](https://peter.bourgon.org), which allows for using either flags or environment variables. Going forward, we'll assume you've chosen flags.

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
        s.loggerChain().
            Append(s.appHandler).
            Append(s.authHandler).
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
