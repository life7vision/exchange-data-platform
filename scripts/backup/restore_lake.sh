#!/usr/bin/env bash
set -euo pipefail

ARCHIVE="${1:-${RESTORE_BACKUP_FILE:-}}"
ROOT="${2:-${LAKE_ROOT:-/opt/exchange-data-platform/lake}}"

if [[ -z "$ARCHIVE" ]]; then
  echo "restore archive path is required" >&2
  exit 1
fi

mkdir -p "$(dirname "$ROOT")"
rm -rf "$ROOT"
tar -C "$(dirname "$ROOT")" -xzf "$ARCHIVE"
