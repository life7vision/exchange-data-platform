#!/usr/bin/env bash
set -euo pipefail

# Backup verification script for Exchange Data Platform
# Verifies backup integrity and tests restore capability

ROOT="${1:-${LAKE_ROOT:-/opt/exchange-data-platform/lake}}"
BACKUP_ROOT="${BACKUP_ROOT:-/opt/exchange-data-platform/backups}"
TEST_DIR="${TEST_DIR:-/tmp/exchange-backup-test-$$}"
LOG_FILE="${LOG_FILE:-${BACKUP_ROOT}/verification-$(date -u +%Y%m%dT%H%M%SZ).log}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
  local level=$1
  shift
  local message="$@"
  local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
  echo "[${timestamp}] [${level}] ${message}" | tee -a "$LOG_FILE"
}

log_success() {
  echo -e "${GREEN}✓${NC} $@" | tee -a "$LOG_FILE"
}

log_error() {
  echo -e "${RED}✗${NC} $@" | tee -a "$LOG_FILE"
}

log_warning() {
  echo -e "${YELLOW}⚠${NC} $@" | tee -a "$LOG_FILE"
}

cleanup() {
  if [[ -d "$TEST_DIR" ]]; then
    log INFO "Cleaning up test directory: $TEST_DIR"
    rm -rf "$TEST_DIR"
  fi
}

trap cleanup EXIT

main() {
  log INFO "Starting backup verification"
  log INFO "Backup root: $BACKUP_ROOT"
  log INFO "Test directory: $TEST_DIR"

  # Check if backup directory exists
  if [[ ! -d "$BACKUP_ROOT" ]]; then
    log_error "Backup directory not found: $BACKUP_ROOT"
    exit 1
  fi

  # Find latest backup
  local latest_backup=$(find "$BACKUP_ROOT" -name "*.tar.gz" -type f | sort -r | head -1)
  if [[ -z "$latest_backup" ]]; then
    log_error "No backup files found in $BACKUP_ROOT"
    exit 1
  fi

  log INFO "Latest backup: $(basename "$latest_backup")"
  log INFO "Backup size: $(du -h "$latest_backup" | cut -f1)"
  log INFO "Backup timestamp: $(stat -c %y "$latest_backup" 2>/dev/null || stat -f %Sm "$latest_backup" 2>/dev/null)"

  # Verify tar.gz integrity
  log INFO "Verifying tar.gz integrity..."
  if ! tar -tzf "$latest_backup" > /dev/null 2>&1; then
    log_error "Backup file is corrupted: $latest_backup"
    exit 1
  fi
  log_success "Backup file integrity verified"

  # Test extraction
  log INFO "Testing backup extraction..."
  mkdir -p "$TEST_DIR"
  if ! tar -xzf "$latest_backup" -C "$TEST_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log_error "Failed to extract backup"
    exit 1
  fi
  log_success "Backup extraction successful"

  # Verify extracted contents
  log INFO "Verifying extracted contents..."
  local extracted_lake="$TEST_DIR/$(basename $ROOT)"
  if [[ ! -d "$extracted_lake" ]]; then
    log_error "Lake directory not found in extracted backup"
    exit 1
  fi
  log_success "Lake directory structure verified"

  # Check for expected subdirectories
  local expected_dirs=("standardized" "temp" "manifests" "rejects" "quality")
  for dir in "${expected_dirs[@]}"; do
    if [[ ! -d "$extracted_lake/$dir" ]]; then
      log_warning "Expected subdirectory not found: $dir"
    else
      local file_count=$(find "$extracted_lake/$dir" -type f 2>/dev/null | wc -l)
      log INFO "Directory $dir contains $file_count files"
    fi
  done

  # Calculate statistics
  local total_files=$(find "$extracted_lake" -type f 2>/dev/null | wc -l)
  local total_size=$(du -sb "$extracted_lake" 2>/dev/null | cut -f1)
  local human_size=$(numfmt --to=iec-i --suffix=B "$total_size" 2>/dev/null || du -sh "$extracted_lake" 2>/dev/null | cut -f1)

  log INFO "Backup statistics:"
  log INFO "  Total files: $total_files"
  log INFO "  Total size: $human_size"

  # Verify file permissions
  log INFO "Verifying file permissions..."
  local permission_issues=0
  while IFS= read -r file; do
    local perms=$(stat -c %a "$file" 2>/dev/null || stat -f %A "$file" 2>/dev/null)
    # Check for world-readable or world-writable (should not exist in backups)
    if [[ "$perms" =~ [24]$ ]]; then
      permission_issues=$((permission_issues + 1))
      log_warning "Unexpected world permissions on: $file ($perms)"
    fi
  done < <(find "$extracted_lake" -type f -not -path "*/.*" 2>/dev/null | head -100)

  if [[ $permission_issues -eq 0 ]]; then
    log_success "No permission issues found (checked first 100 files)"
  else
    log_warning "Found $permission_issues files with potential permission issues"
  fi

  # Manifest validation
  if [[ -d "$extracted_lake/manifests" ]]; then
    log INFO "Validating manifest files..."
    local manifest_count=$(find "$extracted_lake/manifests" -name "*.json" | wc -l)
    if [[ $manifest_count -gt 0 ]]; then
      # Check first manifest for validity
      local sample_manifest=$(find "$extracted_lake/manifests" -name "*.json" | head -1)
      if command -v jq &> /dev/null; then
        if jq empty "$sample_manifest" 2>&1; then
          log_success "Sample manifest is valid JSON"
        else
          log_warning "Sample manifest may have JSON issues"
        fi
      else
        log INFO "jq not available, skipping JSON validation"
      fi
    fi
  fi

  # Create verification report
  local report_file="$LOG_FILE.report"
  {
    echo "=== Backup Verification Report ==="
    echo "Date: $(date -u)"
    echo "Backup file: $(basename "$latest_backup")"
    echo "Backup size: $(du -h "$latest_backup" | cut -f1)"
    echo "Status: VERIFIED"
    echo ""
    echo "Verification Results:"
    echo "  ✓ Tar.gz integrity: PASSED"
    echo "  ✓ Extraction: PASSED"
    echo "  ✓ Directory structure: PASSED"
    echo "  ✓ Total files: $total_files"
    echo "  ✓ Total size: $human_size"
    echo ""
    echo "Recommendations:"
    echo "  - Store backup in secure location"
    echo "  - Consider off-site replication"
    echo "  - Test restore procedure at least monthly"
    echo "  - Document backup recovery time (RTO)"
  } | tee "$report_file"

  log INFO "Backup verification completed successfully"
  log_success "Verification report: $report_file"
  
  return 0
}

main "$@"
