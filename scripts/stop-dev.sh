#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Stopping dev processes..."

# Kill any running backend/frontend dev processes
pkill -f "go run cmd/server/main.go" 2>/dev/null || true
pkill -f "npm run dev" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true

# Stop Docker Compose services
echo "Stopping Docker Compose services..."
docker compose -f "$PROJECT_ROOT/docker-compose.yml" down

echo "Dev environment stopped."
