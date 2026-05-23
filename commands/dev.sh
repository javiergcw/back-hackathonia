#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

set -a
source .env.dev
set +a

go run main.go
