# Disaster Recovery Runbook

Procedures for recovering from various failure scenarios.

## SLOs & RPO/RTO

**Recovery Time Objective (RTO):** 15 minutes  
**Recovery Point Objective (RPO):** 5 minutes

## Common Failure Scenarios

### Scenario 1: Single Worker Crash

**Symptoms:**
- Health check failing for one exchange (e.g., /healthz returns 503)
- No data from that exchange in manifests for >5 minutes
- Alert: "worker_health_status == 0 for exchange=binance"

**Resolution (5 minutes):**
```bash
# Check logs
docker compose logs binance-worker --tail=50

# Restart the worker
docker compose restart binance-worker

# Verify health
curl http://localhost:8080/healthz

# Check data flow
ls -lrt lake/manifests/binance/ | tail -10
```

**Post-incident:**
- Review logs for root cause
- Check for resource constraints
- Adjust memory/CPU limits if needed
- Update dashboards to alert earlier

---

### Scenario 2: All Workers Down

**Symptoms:**
- All health checks failing
- Alert: "Worker fleet health < 50%"
- No data in pipeline for >10 minutes

**Resolution (10-15 minutes):**

```bash
# 1. Check system resources
docker stats

# 2. Check disk space
df -h

# 3. If out of disk, run retention cleanup
bash scripts/retention/purge_old_data.sh

# 4. Restart Docker daemon if needed
sudo systemctl restart docker

# 5. Restart all services
docker compose restart

# 6. Verify all healthy
docker compose ps
for i in 8080 8081 8082 8083; do
  curl -s http://localhost:$i/healthz | jq .
done
```

**If still failing:**
- Check exchange API connectivity: `curl -I https://api.binance.com`
- Review error logs for API rate limits
- Check network firewall rules
- Consider temporary manual restart or failover to backup system

**Post-incident:**
- Implement resource monitoring alerts
- Pre-allocate disk space
- Test automated recovery scripts

---

### Scenario 3: Data Pipeline Stuck

**Symptoms:**
- Services running but no new manifests created
- Alert: "last_success_sync > 10 minutes"
- CPU/memory normal

**Resolution (10 minutes):**

```bash
# 1. Check if data is being fetched
docker compose logs binance-worker --tail=100 | grep -i "fetch"

# 2. Check manifest timestamps
ls -lt lake/manifests/binance/*/latest_manifest.json | head -5

# 3. Check for errors in quality pipeline
ls -lrt lake/quality/ | tail -10

# 4. Manual trigger of sync
# Navigate to health endpoint and trigger run-once
curl -X POST http://localhost:8080/run-once

# 5. Monitor logs
docker compose logs -f binance-worker
```

**If no data:**
- Check API connectivity
- Review API rate limits (may be throttled)
- Check configuration for enabled datasets
- Try reducing batch size temporarily

---

### Scenario 4: Backup Failure

**Symptoms:**
- Alert: "Backup job failed"
- No new backup files in /opt/exchange-data-platform/backups/
- Last backup > 24 hours old

**Resolution (5-10 minutes):**

```bash
# 1. Check backup directory
ls -lrt /opt/exchange-data-platform/backups/ | tail -10

# 2. Check disk space
df -h /opt/exchange-data-platform

# 3. Verify lake directory exists
ls -la /opt/exchange-data-platform/lake/

# 4. Run backup manually
bash scripts/backup/backup_lake.sh

# 5. Verify backup
bash scripts/backup/verify_backup.sh
```

**If backup still fails:**
- Check file permissions: `ls -la lake/`
- Check tar utility: `tar --version`
- Try incremental backup approach
- Consider cloud backup (S3, GCS, etc.)

---

### Scenario 5: Data Corruption in Parquet Files

**Symptoms:**
- Rejects increasing rapidly
- Alert: "error_rate > 5%"
- Parquet read failures in logs

**Resolution (15-30 minutes):**

```bash
# 1. Identify corrupted batch
docker compose logs --since 30m | grep -i "parquet\|error"

# 2. Find corrupted files
find lake/standardized -name "*.parquet" -exec file {} \; | grep -v "Parquet"

# 3. Move corrupted files to quarantine
mkdir -p lake/quarantine
mv lake/standardized/corrupted_file.parquet lake/quarantine/

# 4. Restore from backup (if recent enough)
bash scripts/backup/restore_lake.sh /path/to/backup.tar.gz

# 5. Restart workers
docker compose restart
```

**Prevention:**
- Implement data validation before Parquet write
- Add checksums to manifest entries
- Regular integrity checks on archived data

---

### Scenario 6: Out of Disk Space

**Symptoms:**
- Alert: "disk_usage > 90%"
- Workers slowing down or crashing
- New data stops being written

**Resolution (5-10 minutes):**

```bash
# 1. Check disk usage
df -h /opt/exchange-data-platform

# 2. Identify large directories
du -sh /opt/exchange-data-platform/lake/*
du -sh /opt/exchange-data-platform/backups/*

# 3. Run retention immediately
bash scripts/retention/purge_old_data.sh

# 4. Compress old backups
gzip -v /opt/exchange-data-platform/backups/*.tar

# 5. Move archives to external storage
# tar --to-stdout lake/standardized | gzip | aws s3 cp - s3://backup-bucket/$(date +%s).tar.gz

# 6. Monitor recovery
df -h
du -sh /opt/exchange-data-platform/lake/
```

**Long-term:**
- Implement daily retention cleanup
- Use external storage for backups
- Monitor disk usage trends
- Plan capacity expansion

---

### Scenario 7: Prometheus Metrics Lost

**Symptoms:**
- Alert: "Prometheus not scraping"
- Empty graphs in Grafana
- /metrics endpoint responds but Prometheus down

**Resolution (5 minutes):**

```bash
# 1. Check Prometheus status (if running)
docker compose logs prometheus 2>/dev/null || echo "Not running in compose"

# 2. Verify /metrics endpoint working
curl http://localhost:8080/metrics | head -20

# 3. Restart metrics collection
docker compose restart prometheus 2>/dev/null || echo "Restart manually"

# 4. Check scrape targets
curl http://localhost:9090/api/v1/targets 2>/dev/null | jq '.data.activeTargets'
```

**Post-incident:**
- Historical metrics lost but future data continues
- Implement persistent volume for metrics storage
- Set up remote storage (remote write to external Prometheus)

---

## Recovery Runbooks

### Full System Recovery from Backup

**RTO: 30 minutes | RPO: 24 hours**

```bash
# 1. Backup current corrupted system (if investigating)
bash scripts/backup/backup_lake.sh /opt/exchange-backup-corrupted-$(date +%s).tar.gz

# 2. Stop all services
sudo systemctl stop exchange-data-platform

# 3. Restore lake from latest backup
bash scripts/backup/restore_lake.sh

# 4. Verify restoration
bash scripts/backup/verify_backup.sh

# 5. Start services
sudo systemctl start exchange-data-platform

# 6. Monitor health
watch -n 5 'docker compose ps && curl -s http://localhost:8080/healthz | jq .'
```

### Failover to Secondary Instance

**RTO: 10 minutes | RPO: 5 minutes**

```bash
# On secondary instance:
# 1. Update DNS/load balancer to point to secondary
# 2. Start services
docker compose up -d

# 3. Restore from latest backup
bash scripts/backup/restore_lake.sh

# 4. Verify health
docker compose logs -f

# 5. Resume from last checkpoint
# Workers will auto-resume from checkpoints
```

---

## Testing & Validation

### Monthly DR Test

Schedule monthly tests:
```bash
# 1. Backup current state
bash scripts/backup/backup_lake.sh /tmp/test-backup.tar.gz

# 2. Extract to test directory
mkdir -p /tmp/dr-test
tar -xzf /tmp/test-backup.tar.gz -C /tmp/dr-test

# 3. Verify data integrity
bash scripts/backup/verify_backup.sh /tmp/dr-test

# 4. Test restore on alternate system
# (or container, or VM)

# 5. Document results
echo "DR Test $(date): PASSED" >> /var/log/dr-tests.log
```

### Recovery Metrics
- Track RTO and RPO for each scenario
- Document actual recovery time vs. target
- Update procedures based on learnings
- Conduct quarterly DR drills

---

## Alerting & Escalation

### Alert Thresholds
| Alert | Threshold | Action | Escalation |
|-------|-----------|--------|-----------|
| Worker down | Any unhealthy for 2min | Page on-call | If 2+ workers down, page manager |
| Data lag | >10 min | Check logs | If >30 min, declare incident |
| Error rate | >5% | Investigate | If >10%, rollback deployment |
| Disk usage | >80% | Run cleanup | If >95%, declare SEV1 incident |
| Backup failed | After 1 retry | Alert ops | If 24+ hours old, SEV1 incident |

### Escalation Path
1. On-call engineer (5 min)
2. Team lead (10 min)
3. Engineering manager (15 min)
4. Director (20 min)

---

## Contact Information

- **On-Call**: [Slack channel] or [Phone]
- **Incident Commander**: [Email/Phone]
- **Manager on Duty**: [Email/Phone]
- **Vendor Support**: [Exchange API support links]

---

## Post-Incident

After any incident:
1. Declare incident resolved when:
   - All services healthy
   - Data flowing normally
   - Monitoring shows normal patterns
2. Create incident report within 24 hours
3. Schedule blameless post-mortem
4. Implement preventive measures
5. Update runbooks based on learnings

---

Last Updated: 2026-04-02
