####################################################################
# Builder Stage                                                    #
####################################################################
# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:alpine AS builder

# Create WORKDIR using project's root directory
WORKDIR /go/src/github.com/gilcrest/go-API-template

# Copy the local package files to the container's workspace 
# in the above created WORKDIR
ADD . .

# Build the go-API-template command inside the container
RUN cd cmd/server && go build -o userServer


#####################################################################
# Final Stage                                                       #
#####################################################################
# Pull golang alpine image (very small image, with minimum needed to run Go)
FROM alpine

# Create WORKDIR
WORKDIR /src/github.com/gilcrest/go-API-template/input

# Copy json file needed for feature flags to directory expected by app
# File is copied from the Builder stage image
COPY --from=builder /go/src/github.com/gilcrest/go-API-template/input/httpLogOpt.json .

# Create WORKDIR
WORKDIR /app

# Copy app binary from the Builder stage image
COPY --from=builder /go/src/github.com/gilcrest/go-API-template/cmd/server/userServer .

# Run the userServer command by default when the container starts.
ENTRYPOINT ./userServer

# Document that the service uses port 8080
EXPOSE 8080

# Document that the service uses port 5432
EXPOSE 5432
