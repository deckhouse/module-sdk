# Start from golang base image --platform=linux/amd64
FROM golang:alpine AS builder

RUN set -ex && \
    apk add --no-cache gcc musl-dev

RUN set -ex && \
    rm -f /usr/libexec/gcc/x86_64-alpine-linux-musl/6.4.0/cc1obj && \
    rm -f /usr/libexec/gcc/x86_64-alpine-linux-musl/6.4.0/lto1 && \
    rm -f /usr/libexec/gcc/x86_64-alpine-linux-musl/6.4.0/lto-wrapper && \
    rm -f /usr/bin/x86_64-alpine-linux-musl-gcj

# Set the current working directory inside the container 
WORKDIR /app

# Copy go mod and sum files 
COPY go.mod ./go.mod
COPY go.sum ./go.sum

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed 
RUN /usr/local/go/bin/go mod download 

# Copy the source from the current directory to the working Directory inside the container 
COPY ./module-name ./module-name
COPY ./cmd/hooks/main.go ./cmd/hooks/main.go

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -a -installsuffix cgo -o /app/my-module-hooks ./cmd/hooks/

RUN /app/my-module-hooks hook dump
RUN /app/my-module-hooks hook config >> config.txt
RUN /app/my-module-hooks hook list

# Start a new stage from scratch
FROM alpine:latest AS final

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/my-module-hooks .
