# Byggsteg
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Kopiera go.mod och go.sum först för att utnyttja Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Kopiera hela projektstrukturen
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY data/ ./data/

# Bygg applikationen
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Körsteg
FROM alpine:latest

WORKDIR /app

# Kopiera den byggda binären från byggsteget
COPY --from=builder /app/main .

# Kopiera nödvändiga mappar och filer
COPY internal/templates ./internal/templates
COPY data ./data

EXPOSE 8080

CMD ["./main"] 