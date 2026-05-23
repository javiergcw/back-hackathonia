# Build context: raíz del repo hackathon-bqia
#
#   docker compose up --build
#
# Requiere la red y PostgreSQL del stack banco:
#   docker compose -f ../BACKEND\ 2026/biofood-solution/postgres/docker-compose.banco.yml up -d

FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -ldflags="-w -s" -o /out/main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /out/main .
COPY .env.docker .env.docker

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

ENV PORT=8080
ENV ENV_FILE=.env.docker

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
