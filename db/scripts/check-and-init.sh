#!/usr/bin/env sh
set -euo pipefail
set -x

INIT_FILE="/migrations/000001_init_schema.up.sql"
MARK_TABLE="users._init_done"

echo "== Env =="
echo "PGHOST=$PGHOST PGPORT=$PGPORT PGUSER=$PGUSER PGDATABASE=$PGDATABASE"

echo "== Where am I =="
pwd
ls -la /
ls -la /migrations || true
ls -la /scripts || true

echo "== Wait for Postgres =="
until pg_isready -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" >/dev/null 2>&1; do
  sleep 1
done
echo "Postgres is ready."

echo "== Check marker table =="
if psql -qtAX -c "select to_regclass('$MARK_TABLE') is not null;" | grep -q '^t$'; then
  echo "Marker table $MARK_TABLE exists. Nothing to do."
  exit 0
fi

echo "== Count user tables in users (exclude system) =="
TABLES_COUNT=$(psql -qtAX -c "
  select count(*)
  from pg_catalog.pg_tables
  where schemaname='users'
    and tablename not like 'pg_%'
    and tablename not like 'sql_%';
")
echo "User tables in users: $TABLES_COUNT"

if [ "$TABLES_COUNT" -eq 0 ]; then
  echo "No user tables -> applying INIT: $INIT_FILE"
  if [ ! -f "$INIT_FILE" ]; then
    echo "ERROR: $INIT_FILE not found!" >&2
    exit 1
  fi
  psql --set ON_ERROR_STOP=1 -f "$INIT_FILE"

  echo "Create marker so we don't re-run next time..."
  psql -qtAX -c "create table if not exists $MARK_TABLE(id int primary key default 1);"
  echo "Init applied."
else
  echo "Tables already exist -> skip init."
fi

echo "== Verify some objects =="
psql -c "\dt+"
