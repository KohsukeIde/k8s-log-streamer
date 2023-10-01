# Build stage
FROM golang:1.19-alpine AS builder

# Set the working directory
WORKDIR /go/src/app

# Copy the Go source files into the image
COPY . .

# Download dependencies
RUN go mod download

# Build the Go app to a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o log-streamer .

# Final stage
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /go/src/app/log-streamer /usr/local/bin/log-streamer

# Set the binary as the default command to run
ENTRYPOINT ["/usr/local/bin/log-streamer"]
