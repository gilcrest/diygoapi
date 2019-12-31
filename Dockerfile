####################################################################
# Builder Stage                                                    #
####################################################################
# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.13 as builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server

####################################################################
# Final Stage                                                      #
####################################################################
# Use the official Alpine image for a lean production container.
# https://hub.docker.com/_/alpine
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine:3

# Install ca-certificates bundle inside the docker image
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /server

# Run the web service on container startup.
CMD ["/server", "-datastore=gcp"]
