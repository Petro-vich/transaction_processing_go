FROM golang:1.23.2-alpine AS builder
RUN apk add --no-cache build-base sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o transaction-service ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache sqlite-libs

WORKDIR /app
COPY --from=builder /app/transaction-service .
COPY config/docker.yaml ./config/docker.yaml

RUN mkdir -p storage/sqlite

EXPOSE 8080
CMD ["/app/transaction-service"]
