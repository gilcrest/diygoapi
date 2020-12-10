#!/bin/bash

# Service Name
SERVICE="go-api-basic"

# Google Cloud project ID
PROJECT_ID="go-api-basic-project"

# Image ID
IMAGE="qa"

# Build Tag for Container Registry
TAG="gcr.io/${PROJECT_ID}/${IMAGE}"

# Google DB Instance Name
INSTANCE="go-api-123456:us-east1:go-api-db"

# Database Config for Environment Variables
DBNAME="go-api-basic"
USERNAME="postgres"
PASSWORD="fakepassword"
HOST="/cloudsql/${INSTANCE}"
PORT="5432"

# Sendgrid API key for email delivery
SENDGRID='SG.allfake--FTf'

echo "---------------------------------------------"
echo "About to build container to Google Cloud Registry with tag $TAG"
echo "---------------------------------------------"
gcloud builds submit --tag "${TAG}"

echo "---------------------------------------------"
echo "About to deploy image to Google Cloud Run with real DB connection"
echo "---------------------------------------------"
gcloud run deploy "${SERVICE}" --image "${TAG}" --platform managed \
--allow-unauthenticated \
--add-cloudsql-instances "${INSTANCE}" \
--set-env-vars \
INSTANCE-CONNECTION-NAME="${INSTANCE}",\
PG_GCP_DBNAME="${DBNAME}",\
PG_GCP_USERNAME="${USERNAME}",\
PG_GCP_PASSWORD="${PASSWORD}",\
PG_GCP_HOST="${HOST}",\
PG_GCP_PORT="${PORT}",\
SENDGRID_API_KEY="${SENDGRID}"

echo "---------------------------------------------"
echo "About to add allUsers"
echo "---------------------------------------------"
gcloud run services add-iam-policy-binding "${SERVICE}" \
    --platform managed \
    --member="allUsers" \
    --role="roles/run.invoker"

echo "---------------------------------------------"
echo "Build Script Completed"
echo "---------------------------------------------"

