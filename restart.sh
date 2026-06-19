#!/usr/bin/env bash
# Restart all Docker services (rebuilds backend/frontend if code changed)
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

echo "Stopping containers..."
docker compose down

echo "Rebuilding and starting..."
docker compose up --build -d

echo "Waiting for health checks..."
TRIES=0
until docker compose ps mysql | grep -q "healthy" 2>/dev/null; do
    TRIES=$((TRIES + 1))
    [ "$TRIES" -ge 60 ] && break
    sleep 2
done

echo ""
docker compose ps
echo ""
echo "✔ Services restarted. Logs: docker compose logs -f"
