#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

SERVICE="hackathon-qia-backend"
NETWORK="banco-agent-net"

usage() {
  cat <<EOF
Uso: ./commands/deploy.sh [comando]

Comandos:
  up        Levanta el contenedor (default: docker compose up --build -d)
  down      Detiene y elimina el contenedor (docker compose down)
  restart   down + up
  logs      Ver logs en vivo del backend
  status    Estado del contenedor y healthcheck
  help      Muestra esta ayuda

Ejemplo:
  ./commands/deploy.sh
  ./commands/deploy.sh up
  ./commands/deploy.sh logs
EOF
}

ensure_network() {
  if ! docker network inspect "$NETWORK" >/dev/null 2>&1; then
    echo "→ Creando red Docker: $NETWORK"
    docker network create "$NETWORK"
  fi
}

cmd_up() {
  ensure_network
  echo "→ Levantando $SERVICE..."
  docker compose up --build -d
  echo ""
  docker compose ps

  local app_url="http://localhost:8090"
  if [ -f .env.docker ]; then
    app_url="$(grep '^APP_PUBLIC_URL=' .env.docker | cut -d= -f2- || true)"
    app_url="${app_url:-http://localhost:8090}"
  fi

  echo ""
  echo "→ API: $app_url"
  echo "→ Health: curl $app_url/health"
  echo "→ Webhook: curl $app_url/whatsapp/webhook"
}

cmd_down() {
  echo "→ Deteniendo $SERVICE..."
  docker compose down
}

cmd_restart() {
  cmd_down
  cmd_up
}

cmd_logs() {
  docker compose logs -f -t "$SERVICE"
}

cmd_status() {
  docker compose ps
  echo ""
  docker inspect --format='{{.State.Health.Status}}' "$SERVICE" 2>/dev/null || true
}

case "${1:-up}" in
  up) cmd_up ;;
  down) cmd_down ;;
  restart) cmd_restart ;;
  logs) cmd_logs ;;
  status) cmd_status ;;
  help|-h|--help) usage ;;
  *) echo "Comando desconocido: $1"; echo ""; usage; exit 1 ;;
esac
