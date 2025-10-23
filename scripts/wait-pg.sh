#!/usr/bin/env bash
set -euo pipefail

# usage: ./scripts/wait-pg.sh <compose_service> <db_name> [user]
# examples:
#   ./scripts/wait-pg.sh office-db office
#   ./scripts/wait-pg.sh logistic-db logistic

SERVICE="${1:?compose service required}"
DB="${2:?database name required}"
USER="${3:-postgres}"

echo "Waiting for $SERVICE (db=$DB user=$USER) ..."
# ждём, пока PostgreSQL внутри контейнера будет принимать соединения
until docker compose exec -T "$SERVICE" \
  pg_isready -U "$USER" -d "$DB" -h 127.0.0.1 -p 5432 >/dev/null 2>&1
do
  sleep 1
done
echo "$SERVICE is ready."
