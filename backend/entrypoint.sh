#!/bin/bash
set -e

echo "🚀 Starting Teralux Backend..."

DB_TYPE="${DB_TYPE:-sqlite}"
DB_SQLITE_PATH="${DB_SQLITE_PATH:-./tmp/teralux.db}"

run_sqlite_migrations() {
  mkdir -p "$(dirname "$DB_SQLITE_PATH")"
  local abs_path
  abs_path=$(realpath -m "$DB_SQLITE_PATH")
  echo "🔄 Running SQLite migrations at $abs_path..."
  migrate -path ./migrations -database "sqlite3://$abs_path" up
}

case "$DB_TYPE" in
  sqlite)
    run_sqlite_migrations || { echo "❌ SQLite migrations failed"; exit 1; }
    ;;
  mysql)
    echo "ℹ️  Skipping migrations for MySQL (DB_TYPE=mysql)"
    ;;
  *)
    echo "❌ Unsupported DB_TYPE: $DB_TYPE"
    exit 1
    ;;
esac

# Start the application
echo "🚀 Starting application... (DB_TYPE=$DB_TYPE)"
exec ./main
