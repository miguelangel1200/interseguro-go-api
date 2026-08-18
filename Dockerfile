# ---- Etapa de build ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copia los módulos y descarga dependencias para aprovechar la caché de Docker.
COPY go.mod go.sum ./
RUN go mod download

# Copia el código fuente y compila un binario estático.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-api ./cmd

# ---- Etapa de runtime ----
FROM alpine:3.20

RUN adduser -D -u 1000 appuser
WORKDIR /app

COPY --from=builder /go-api /app/go-api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/go-api"]
