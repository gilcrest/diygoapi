#!/bin/bash
# =============================================================================
# Author: Dan Gillis
# Created On: 28 Mar 2018
# Purpose: Script builds a docker image then runs a container with that image
# =============================================================================
#

# Build image with gilcrest as repository, go-api-template as build name, latest
# as build tag
docker image build -t gilcrest/go-api-template:latest .

# Run container using previously built image (gilcrest/go-api-template)
# expose port 5432 on the host to port 5432 on the container for postgres 
# exposing port 8080 on the host to port 8080 on the container for http access
# load environment variables using the env file
docker container run -d -p 5432:5432 -p 8080:8080 --env-file ./test.env --name user-server gilcrest/go-api-template
# docker container exec -it user-server /bin/bash

#docker container run -it -p 5432:5432 -p 8080:8080 --env-file ./test.env --name user-server gilcrest/go-api-template /bin/bash
