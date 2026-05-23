#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

set -a
source .env.prod
set +a

go build -o bin/server main.go
./bin/server
