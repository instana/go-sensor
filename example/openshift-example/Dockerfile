FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Copy the parent go-sensor module
COPY ../../go.mod ../../go.sum ../../
COPY ../../*.go ../../
COPY ../../acceptor ../../acceptor
COPY ../../autoprofile ../../autoprofile
COPY ../../aws ../../aws
COPY ../../docker ../../docker
COPY ../../gcloud ../../gcloud
COPY ../../internal ../../internal
COPY ../../logger ../../logger
COPY ../../process ../../process
COPY ../../secrets ../../secrets
COPY ../../w3ctrace ../../w3ctrace

# Copy the instaamqp091 instrumentation module
COPY ../../instrumentation/instaamqp091 ../../instrumentation/instaamqp091

# Download dependencies
RUN go mod download

# Copy application source
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o myapp .

FROM alpine:3.21 AS runner

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/myapp .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8085
ENTRYPOINT ["/app/myapp"]

# Made with Bob
