# ---------- Build Stage ----------
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

RUN go test -v ./tests

# IMPORTANT: build from cmd/api
RUN go build -o app ./cmd/api

# ---------- Runtime Stage ----------
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

# Copy compiled binary
COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]