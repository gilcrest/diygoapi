#!/bin/bash
# =============================================================================
# Author: Dan Gillis
# Created On: 28 Mar 2018
# Purpose: Script stops the named user-server container and then removes it and
#          the associated image (gilcrest/go-api-template:latest)
# =============================================================================
#

# Stop container using container name (set with --name during docker run)
docker container stop user-server

# Remove container using container name (set with --name during docker run)
docker container rm user-server

# Remove container using container name (set with --name during docker run)
docker image rm gilcrest/go-api-template:latest
