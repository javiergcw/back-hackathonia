#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
# shellcheck source=_go_env.sh
source "$(dirname "$0")/_go_env.sh"

ENV_FILE="${ENV_FILE:-.env.dev}"

set -a
source "$ENV_FILE"
set +a

echo "→ Migraciones SQL en migrations/ (env: ${ENV_FILE})"
go run ./cmd/migrate
