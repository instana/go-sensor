FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o myapp .

FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/myapp .

EXPOSE 8085
ENTRYPOINT ["/app/myapp"]