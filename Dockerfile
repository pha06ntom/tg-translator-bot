FROM golang:1.24.1-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bot ./cmd/bot

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tesseract-ocr \
    tesseract-ocr-rus \
    tesseract-ocr-eng \
    poppler-utils \
    imagemagick \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/bot /app/bot
COPY migrations /app/migrations

CMD ["/app/bot", "-config", "/app/config.yaml"]