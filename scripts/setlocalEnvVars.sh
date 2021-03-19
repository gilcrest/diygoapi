#!/bin/bash

# In order to run this script from your current shell
# you need to "source" the script, so run as either:
# ". ./setlocalEnvVars.sh" or "source ./setlocalEnvVars.sh"

# Database Environment variables
export DB_NAME="go_api_basic"
export DB_USER="postgres"
export DB_PASSWORD=""
export DB_HOST="localhost"
export DB_PORT="5432"

# Google Oauth2 token - use the Google oauth2 playground https://developers.google.com/oauthplayground/
# and replace below
export GOOGLE_ACCESS_TOKEN="REPLACEME"