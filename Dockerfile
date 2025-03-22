# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang:1.12-alpine base image
FROM golang:1.23-alpine

WORKDIR /app
# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh gcc musl-dev

RUN go install github.com/air-verse/air@latest

# Set the Current Working Directory inside the container

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

ENV CGO_ENABLED=1 \
    GOFLAGS="-buildvcs=false"


# Expose port 8080 to the outside world
EXPOSE 8090

# Run the executable
CMD ["air", "-c", ".air.toml"]