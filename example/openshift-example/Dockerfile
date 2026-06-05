FROM public.ecr.aws/docker/library/golang:1.25 AS builder

WORKDIR /app

# Copy the entire go-sensor source code (since build context is at root)
COPY . .

# Change to the example directory
WORKDIR /app/example/openshift-example

# Tidy dependencies to update go.sum with replace directives
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

FROM alpine:3.21 AS runner

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/example/openshift-example/myapp .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8085
ENTRYPOINT ["/app/myapp"]

# Made with Bob
