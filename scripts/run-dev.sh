#!/usr/bin/env bash
set -euo pipefail

# ── Config ────────────────────────────────────────────────
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PID=""
FRONTEND_PID=""

# ── Cleanup ───────────────────────────────────────────────
cleanup() {
  echo ""
  echo "Shutting down dev environment..."
  [[ -n "$BACKEND_PID" ]] && kill "$BACKEND_PID" 2>/dev/null || true
  [[ -n "$FRONTEND_PID" ]] && kill "$FRONTEND_PID" 2>/dev/null || true
  wait 2>/dev/null || true
  echo "Done."
}
trap cleanup EXIT INT TERM

# ── Pre-flight ────────────────────────────────────────────
if ! command -v docker &>/dev/null; then
  echo "Error: docker is not installed or not in PATH" >&2
  exit 1
fi

if ! docker info &>/dev/null; then
  echo "Error: docker daemon is not running" >&2
  exit 1
fi

# ── Docker Compose services ───────────────────────────────
echo "Starting Docker Compose services..."
docker compose -f "$PROJECT_ROOT/docker-compose.yml" up -d

# ── Wait for PostgreSQL ───────────────────────────────────
echo "Waiting for PostgreSQL to be ready..."
RETRIES=30
until docker compose -f "$PROJECT_ROOT/docker-compose.yml" exec -T postgres pg_isready -U cnow -d cnow &>/dev/null; do
  RETRIES=$((RETRIES - 1))
  if [[ $RETRIES -le 0 ]]; then
    echo "Error: PostgreSQL did not become ready in time" >&2
    exit 1
  fi
  echo "  PostgreSQL not ready yet, retrying... ($RETRIES attempts left)"
  sleep 2
done
echo "PostgreSQL is ready."

# ── Export DB env vars ────────────────────────────────────
export CNOW_DB_HOST="${CNOW_DB_HOST:-localhost}"
export CNOW_DB_PORT="${CNOW_DB_PORT:-5432}"
export CNOW_DB_USER="${CNOW_DB_USER:-cnow}"
export CNOW_DB_PASSWORD="${CNOW_DB_PASSWORD:-cnow}"
export CNOW_DB_NAME="${CNOW_DB_NAME:-cnow}"
export CNOW_DB_SSLMODE="${CNOW_DB_SSLMODE:-disable}"
export CNOW_HTTP_ADDR="${CNOW_HTTP_ADDR:-:8080}"
export CNOW_ENV="${CNOW_ENV:-dev}"

# Load .env if present
if [[ -f "$PROJECT_ROOT/.env" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "$PROJECT_ROOT/.env"
  set +a
fi

# ── Start backend ─────────────────────────────────────────
echo "Starting backend..."
cd "$PROJECT_ROOT/backend"
go run cmd/server/main.go &
BACKEND_PID=$!
cd "$PROJECT_ROOT"

# ── Start frontend ────────────────────────────────────────
echo "Starting frontend..."
cd "$PROJECT_ROOT/frontend"
npm run dev &
FRONTEND_PID=$!
cd "$PROJECT_ROOT"

echo ""
echo "Dev environment is running!"
echo "  Backend:  http://localhost:8080"
echo "  Frontend: http://localhost:5173"
echo "  Postgres: localhost:5432"
echo "  Redis:    localhost:6379"
echo "  Temporal: localhost:7233"
echo ""
echo "Press Ctrl+C to stop."

# Wait for any child to exit
wait -n "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
