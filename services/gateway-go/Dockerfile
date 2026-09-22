# Etapa 1: Compilación
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api-gateway ./cmd/api/main.go

# Etapa 2: Imagen ejecutable ultra ligera
FROM alpine:latest
WORKDIR /app
COPY --from=builder /api-gateway .

EXPOSE 8080
CMD ["./api-gateway"]