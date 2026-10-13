#!/usr/bin/env bash
set -euo pipefail

root_dir=$(cd "$(dirname "$0")/.." && pwd)
schema_file="$root_dir/script/schema/v0.1.0.sql"

if [[ ! -f "$schema_file" ]]; then
  echo "schema file not found: $schema_file" >&2
  exit 1
fi

run_psql() {
  if [[ -n "${GOSILITE_SCHEMA_DSN:-}" ]]; then
    if ! command -v psql >/dev/null 2>&1; then
      echo "psql is required when GOSILITE_SCHEMA_DSN is set" >&2
      exit 1
    fi
    psql "$GOSILITE_SCHEMA_DSN" "$@"
    return
  fi

  docker exec -i "${GOSILITE_POSTGRES_CONTAINER:-gosilite-postgres}" \
    psql -U "${GOSILITE_POSTGRES_USER:-gosilite}" -d "${GOSILITE_POSTGRES_DB:-gosilite}" "$@"
}

existing_table=$(run_psql -Atqc "SELECT to_regclass('public.system_user') IS NOT NULL;" | tr -d '[:space:]')
if [[ "$existing_table" != "f" ]]; then
  echo "refusing to initialize a non-empty database; back up and recreate it before applying v0.1.0" >&2
  exit 1
fi

run_psql -v ON_ERROR_STOP=1 < "$schema_file"
echo "initialized v0.1.0 schema"
