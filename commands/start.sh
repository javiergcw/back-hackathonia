#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

export ENV_FILE=".env.prod"
go build -o bin/server main.go
./bin/server
