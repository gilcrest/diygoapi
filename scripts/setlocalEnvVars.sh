#!/bin/bash

# In order to run this script from your current shell
# you need to "source" the script, so run as either:
# ". ./setlocalEnvVars.sh" or "source ./setlocalEnvVars.sh"

# Database Environment variables
export PG_APP_DBNAME="go_api_basic"
export PG_APP_USERNAME="postgres"
export PG_APP_PASSWORD=""
export PG_APP_HOST="localhost"
export PG_APP_PORT="5432"

# Google Oauth2 token - use the Google oauth2 playground https://developers.google.com/oauthplayground/
# and replace below
export GOOGLE_ACCESS_TOKEN="REPLACEME"