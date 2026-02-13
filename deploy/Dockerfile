####################################################################
# Builder Stage                                                    #
####################################################################
# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:latest as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy the local code files and directories to the container's workspace
# in the above created WORKDIR
COPY . ./

# Build the binary inside the container
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o srvr

####################################################################
# Final Stage                                                      #
####################################################################
# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:latest

# Install ca-certificates bundle inside the docker image
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/srvr /srvr

# Get timezone zip from go library and add environment variable to point to it
ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

# Run the web service on container startup.
CMD ["/srvr", "-log-level=debug"]
