# Production Deployment Checklist

Pre-deployment verification checklist for Exchange Data Platform.

## Pre-Deployment (48 hours before)

### Infrastructure
- [ ] Server capacity verified (CPU, RAM, disk)
- [ ] Network connectivity tested (DNS, external APIs)
- [ ] Firewall rules configured
  - [ ] Inbound: Allow health/metrics from monitoring
  - [ ] Outbound: Allow to exchange APIs (Binance, Bybit)
- [ ] Storage backends available and responsive
- [ ] Time synchronization verified on all nodes

### Security
- [ ] Secrets loaded into vault/secret manager
  - [ ] Binance API keys rotated
  - [ ] Bybit API keys rotated
  - [ ] Database credentials secure
- [ ] SSL/TLS certificates valid and installed
- [ ] SSH keys distributed to ops team
- [ ] Audit logging configured
- [ ] Network policies and firewall rules deployed

### Data & Backups
- [ ] Latest backup verified and tested
- [ ] Restore procedure tested successfully
- [ ] Backup retention policies configured
- [ ] Backup encryption keys secured
- [ ] Lake directory permissions correct (755)
- [ ] Old data purged per retention policy

### Configuration
- [ ] Environment files (.env) configured per deployment
- [ ] All exchange configs verified
- [ ] Job configurations reviewed
- [ ] Logging level set appropriately
- [ ] Resource limits configured:
  - [ ] Development: 1 CPU / 512MB per worker
  - [ ] Production: 2 CPU / 1GB per worker

### Monitoring & Alerting
- [ ] Prometheus configured and scraping
- [ ] Grafana dashboards created
- [ ] Alert rules defined and tested
  - [ ] High error rate (>5%)
  - [ ] Data lag (>5 min without sync)
  - [ ] Service restart loops
  - [ ] Disk usage >80%
  - [ ] Backup failures
- [ ] Centralized logging configured (ELK/Datadog/etc)
- [ ] On-call rotation scheduled

### Documentation
- [ ] Runbooks reviewed and printed/accessible
- [ ] Contact list updated
- [ ] Disaster recovery plan reviewed
- [ ] Change log prepared
- [ ] Rollback procedure documented

## Day of Deployment

### 2 Hours Before
- [ ] System status check - all services healthy
- [ ] Backup completed successfully
- [ ] Monitoring dashboard accessible
- [ ] On-call engineer standing by
- [ ] Deployment plan reviewed with team

### Deployment Window
- [ ] Pull latest code from main branch
- [ ] Build Docker images
  ```bash
  docker compose -f deployments/docker/compose.yml \
    -f deployments/docker/compose.prod.yml build
  ```
- [ ] Run pre-deployment tests
  ```bash
  go test ./... -v
  ```
- [ ] Stop current services (if applicable)
  ```bash
  docker compose -f deployments/docker/compose.yml \
    -f deployments/docker/compose.prod.yml down
  ```
- [ ] Start new services
  ```bash
  docker compose -f deployments/docker/compose.yml \
    -f deployments/docker/compose.prod.yml up -d
  ```
- [ ] Wait for health checks to pass (30 seconds)
- [ ] Verify all services running
  ```bash
  docker compose -f deployments/docker/compose.yml \
    -f deployments/docker/compose.prod.yml ps
  ```

### Post-Deployment (1 hour)
- [ ] All services healthy (status check)
- [ ] Metrics flowing into Prometheus
- [ ] No errors in logs
  ```bash
  docker compose -f deployments/docker/compose.yml \
    -f deployments/docker/compose.prod.yml logs --since 5m
  ```
- [ ] Data flowing through pipeline (check last sync timestamp)
- [ ] Alerts not firing
- [ ] Backup completed successfully

### Post-Deployment (24 hours)
- [ ] Monitor metrics for stability
- [ ] No unusual errors or warnings
- [ ] Data pipeline processing normally
- [ ] Retention cleanup jobs ran successfully
- [ ] Backup completed and verified
- [ ] Disaster recovery test scheduled

## Rollback Plan

If deployment fails, rollback as follows:

1. Stop new services
   ```bash
   docker compose -f deployments/docker/compose.yml \
     -f deployments/docker/compose.prod.yml down
   ```

2. Restore previous version
   ```bash
   git checkout <previous-commit-sha>
   docker compose -f deployments/docker/compose.yml \
     -f deployments/docker/compose.prod.yml build
   docker compose -f deployments/docker/compose.yml \
     -f deployments/docker/compose.prod.yml up -d
   ```

3. Verify health and data integrity
4. Investigate failure
5. Document incident

## Performance Baselines

Expected metrics (adjust based on actual testing):

| Metric | Development | Production |
|--------|-----------|-----------|
| Data latency | <2 min | <1 min |
| API response time | <500ms | <300ms |
| Error rate | <0.1% | <0.01% |
| CPU usage per worker | 15-25% | 30-40% |
| Memory usage per worker | 100-200MB | 300-500MB |
| Parquet write time | <5s | <2s |

## Sign-off

- Deployment Lead: _____________ Date: _______
- Ops Manager: _____________ Date: _______
- Senior Engineer: _____________ Date: _______

---

For more information, see:
- docs/operations/runbook.md - Daily operations
- docs/operations/systemd.md - Service management
- docs/operations/security.md - Security procedures
