# Build context: raíz del repo hackathon-qia
#
#   docker compose up --build
#
# Requiere la red banco-agent-net (ver docker-compose.yml)

FROM golang:1.22-alpine AS builder

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
COPY data/ data/

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

ENV PORT=8080
ENV ENV_FILE=.env.docker

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 -O /dev/null http://localhost:8080/health || exit 1

CMD ["./main"]