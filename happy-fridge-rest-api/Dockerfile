FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server /app/server

COPY migrations ./migrations

COPY configs/config.toml ./configs/config.toml

RUN apk add --no-cache ca-certificates

EXPOSE 8080

CMD ["/app/server", "-config", "/app/configs/config.toml"]