#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

ENV_FILE="${ENV_FILE:-.env.dev}"

set -a
source "$ENV_FILE"
set +a

go run ./cmd/migrate
