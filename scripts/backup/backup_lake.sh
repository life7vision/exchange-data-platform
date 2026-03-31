#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-${LAKE_ROOT:-/opt/exchange-data-platform/lake}}"
BACKUP_ROOT="${BACKUP_ROOT:-/opt/exchange-data-platform/backups}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"

mkdir -p "$BACKUP_ROOT"
ARCHIVE="$BACKUP_ROOT/exchange-data-platform-lake-$STAMP.tar.gz"

tar -C "$(dirname "$ROOT")" -czf "$ARCHIVE" "$(basename "$ROOT")"
echo "$ARCHIVE"
