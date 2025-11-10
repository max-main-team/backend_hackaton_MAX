#!/usr/bin/env bash
set -euo pipefail

# ---- Настройка --------------------------------------------------------------
# Какие схемы и таблицы считаем «признаком инициализации»
REQUIRED_SCHEMAS=("users")
REQUIRED_TABLES=("users.max_users_data")

INIT_FILE="/migrations/000001_init_schema.up.sql"

# ---- Функции ----------------------------------------------------------------
function wait_pg() {
  echo "Waiting for Postgres at ${PGHOST}:${PGPORT}..."
  until pg_isready -h "${PGHOST}" -p "${PGPORT}" -U "${PGUSER}" -d "${PGDATABASE}" >/dev/null 2>&1; do
    sleep 1
  done
  echo "Postgres is ready."
}

function schema_exists() {
  local schema="$1"
  psql -qtAX -c "select 1 from information_schema.schemata where schema_name='${schema}' limit 1;" | grep -q '^1$'
}

function table_exists() {
  local schema_table="$1"
  # to_regclass возвращает NULL, если таблицы нет
  psql -qtAX -c "select to_regclass('${schema_table}') is not null;" | grep -q '^t$'
}

# ---- Логика -----------------------------------------------------------------
wait_pg

missing=0

# Проверка схем
for s in "${REQUIRED_SCHEMAS[@]}"; do
  if schema_exists "$s"; then
    echo "✔ schema '${s}' exists"
  else
    echo "✖ schema '${s}' is missing"
    missing=1
  fi
done

# Проверка таблиц
for t in "${REQUIRED_TABLES[@]}"; do
  if table_exists "$t"; then
    echo "✔ table '${t}' exists"
  else
    echo "✖ table '${t}' is missing"
    missing=1
  fi
done

if [[ "${missing}" -eq 0 ]]; then
  echo "Nothing to do: all required schemas/tables exist."
  exit 0
fi

# Применяем init-миграцию
if [[ -f "${INIT_FILE}" ]]; then
  echo "Applying ${INIT_FILE}..."
  psql --set ON_ERROR_STOP=1 -f "${INIT_FILE}"
  echo "Init migration applied successfully."
else
  echo "ERROR: ${INIT_FILE} not found. Cannot initialize schema." >&2
  exit 1
fi

# Доп. верификация после применения
post_missing=0
for s in "${REQUIRED_SCHEMAS[@]}"; do
  schema_exists "$s" || post_missing=1
done
for t in "${REQUIRED_TABLES[@]}"; do
  table_exists "$t" || post_missing=1
done

if [[ "${post_missing}" -eq 0 ]]; then
  echo "Schema verified after init."
else
  echo "WARNING: Some schemas/tables still missing after init." >&2
fi
