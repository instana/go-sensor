FROM public.ecr.aws/docker/library/golang:1.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

FROM public.ecr.aws/docker/library/alpine:3.21.2 AS runner
WORKDIR /app
COPY --from=builder /app/myapp .

EXPOSE 8085
ENTRYPOINT ["/app/myapp"]
