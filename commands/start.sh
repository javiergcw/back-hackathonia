#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
# shellcheck source=_go_env.sh
source "$(dirname "$0")/_go_env.sh"

set -a
source .env.prod
set +a

go build -o bin/server main.go
./bin/server
