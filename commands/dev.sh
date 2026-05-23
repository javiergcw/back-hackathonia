#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

export ENV_FILE=".env.dev"
go run main.go
