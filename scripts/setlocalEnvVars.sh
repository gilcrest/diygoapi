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

# Google Oauth2 token
export GOOGLE_ACCESS_TOKEN="ya29.A0AfH6SMBXbKWdkWYVzzzTw7UC4pSFdbBk9s7Mnb1E-rqOkEZx6vScm8PgUQXVN57BcpqaO7aDR4aCxSo8907gGI9K-PdWxy_lDmvhw-fh4rbhROE6S6o_XefZJD3o0_VvnpEM7rERPGrYvSkjvU0_ygb3eBWx"