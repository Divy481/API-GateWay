# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o gateway ./cmd/gateway

# Run stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/gateway .
COPY config.yaml .

EXPOSE 8080
CMD ["./gateway"]
