#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-${LAKE_ROOT:-/opt/exchange-data-platform/lake}}"
STANDARDIZED_DAYS="${RETENTION_DAYS_STANDARDIZED:-14}"
CURATED_DAYS="${RETENTION_DAYS_CURATED:-30}"
MANIFESTS_DAYS="${RETENTION_DAYS_MANIFESTS:-30}"
REJECTS_DAYS="${RETENTION_DAYS_REJECTS:-7}"
QUALITY_DAYS="${RETENTION_DAYS_QUALITY:-14}"

purge_files() {
  local path="$1"
  local days="$2"
  if [[ -d "$path" ]]; then
    find "$path" -type f -mtime +"$days" -print -delete
  fi
}

purge_empty_dirs() {
  local path="$1"
  if [[ -d "$path" ]]; then
    find "$path" -type d -empty -delete
  fi
}

purge_files "$ROOT/standardized" "$STANDARDIZED_DAYS"
purge_files "$ROOT/curated" "$CURATED_DAYS"
purge_files "$ROOT/manifests" "$MANIFESTS_DAYS"
purge_files "$ROOT/rejects" "$REJECTS_DAYS"
purge_files "$ROOT/quality" "$QUALITY_DAYS"

purge_empty_dirs "$ROOT/standardized"
purge_empty_dirs "$ROOT/curated"
purge_empty_dirs "$ROOT/manifests"
purge_empty_dirs "$ROOT/rejects"
purge_empty_dirs "$ROOT/quality"
