# Build stage
FROM golang:1.19-alpine AS builder

# Set the working directory
WORKDIR /go/src/app

# Copy the Go source files into the image
COPY . .

# Download dependencies
RUN go mod download

# Build the Go app to a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o oom-fetch .

# Final stage
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /go/src/app/oom-fetch /usr/local/bin/oom-fetch

# Set the binary as the default command to run
ENTRYPOINT ["/usr/local/bin/oom-fetch"]

# マルチステージビルドしよう。