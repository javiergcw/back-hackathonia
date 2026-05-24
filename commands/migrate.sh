#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
# shellcheck source=_go_env.sh
source "$(dirname "$0")/_go_env.sh"

ENV_FILE="${ENV_FILE:-.env.dev}"

set -a
source "$ENV_FILE"
set +a

go run ./cmd/migrate
