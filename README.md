# go-API-basic

A RESTful API template (built with Go)

The goal of this project is to make an example/template of a relational database-backed REST HTTP API that has characteristics needed to ensure success in a high volume environment. I'm gearing this towards beginners, as I struggled with a lot of this over the past couple of years and would like to help others getting started.

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

## 12/17/2020 - README under construction

I am close to finished for this phase of code refactor and am currently rewriting this doc... I hope to publish the rewritten readme before Christmas.
