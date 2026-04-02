# Security Best Practices

## Secrets Management

### Development Environment
Use `.env` and `secrets.env` files locally (never commit to git).

```bash
cp deployments/docker/.env.example deployments/docker/.env
cp deployments/docker/secrets.example.env deployments/docker/secrets.env
chmod 600 deployments/docker/secrets.env
```

### Production Environment

**DO NOT use plain .env files in production.**

#### Option 1: Docker Secrets (Recommended for Docker Swarm)
```bash
# Create secrets
echo "your_binance_api_key" | docker secret create binance_api_key -
echo "your_binance_secret" | docker secret create binance_api_secret -

# Reference in compose.yml
services:
  binance-worker:
    secrets:
      - binance_api_key
      - binance_api_secret
    environment:
      BINANCE_API_KEY_FILE: /run/secrets/binance_api_key

secrets:
  binance_api_key:
    external: true
  binance_api_secret:
    external: true
```

#### Option 2: HashiCorp Vault (Recommended for Kubernetes)
```bash
# Vault setup
vault kv put secret/exchange-data-platform/binance \
  api_key="your_api_key" \
  api_secret="your_secret"

# Application loads from Vault
# Requires sidecar or init container
```

#### Option 3: Cloud Provider Secrets Manager
- AWS Secrets Manager
- Google Secret Manager
- Azure Key Vault

Load at runtime via application code or sidecar.

## Credential Rotation

Rotate credentials regularly:
- API keys: Every 30-90 days
- Database passwords: Every 30 days
- Backup encryption keys: When policy requires

Document rotation schedule in runbooks.

## Network Security

### Port Exposure
- Health endpoints (8080-8083): Restrict to monitoring systems
- Metrics endpoint (9090): Restrict to Prometheus/monitoring only
- Never expose directly to internet

### Docker Network
Use internal networks, not host mode:
```yaml
networks:
  internal:
    driver: bridge

services:
  binance-worker:
    networks:
      - internal
```

### TLS/HTTPS
- Use TLS for all external API calls (already done)
- Consider TLS between services for sensitive data
- Use self-signed certs or internal CA for internal services

## Resource Limits

All services have resource limits configured in `compose.yml`:
- **Worker processes**: 1-2 CPU, 512MB-1GB RAM
- **Scheduler**: 0.5 CPU, 256MB RAM
- **Backups**: Should be constrained to prevent disk I/O saturation

Monitor actual usage and adjust based on observed patterns.

## File Permissions

### Lake Directory
```bash
chmod 755 /opt/exchange-data-platform/lake
chmod 755 /opt/exchange-data-platform/backups
chmod 600 /opt/exchange-data-platform/deployments/docker/secrets.env
```

### Backup Archives
```bash
# Ensure backups are not world-readable
chmod 600 /opt/exchange-data-platform/backups/*.tar.gz
```

### Configuration Files
```bash
# Restrict access to config directory
chmod 700 /opt/exchange-data-platform/deployments/docker
chmod 600 /opt/exchange-data-platform/deployments/docker/.env
```

## Logging & Monitoring

### Log Output
- All logs are JSON-formatted and sent to stdout
- In production, aggregate logs to centralized system
- Do NOT log credentials, API keys, or sensitive data
- Use structured logging with appropriate log levels

### Audit Logging
- Log all authentication attempts
- Log configuration changes
- Log backup/restore operations
- Retain audit logs for 90 days minimum

### Health Monitoring
- Monitor worker health checks continuously
- Alert on failed health checks
- Alert on data lag (no successful sync > 5 minutes)
- Alert on high error rates

## Access Control

### Running Services
```bash
# Run as non-root user (use Docker USER directive)
# Current: distroless base runs as UID 1000 (non-root)
```

### Backup Access
```bash
# Limit backup restore access
sudo chown root:backup /opt/exchange-data-platform/backups
sudo chmod 750 /opt/exchange-data-platform/backups
```

### Code Repository
- Use branch protection rules
- Require code review before merge
- Never commit credentials
- Scan for secrets with `git-secrets` or similar

## Dependency Management

- Keep Go dependencies updated
- Run `go mod tidy` and test before updating
- Monitor for CVEs in dependencies
- Use `go mod audit` (when available in your Go version)

## Encryption

### Data at Rest
- Backup files: Use encryption (gpg, openssl)
```bash
# Example: Encrypt backup
gpg --symmetric --cipher-algo AES256 backup-file.tar.gz

# Decrypt
gpg --decrypt backup-file.tar.gz.gpg > backup-file.tar.gz
```

### Data in Transit
- All external API calls use HTTPS
- Consider TLS between internal services
- Use certificate pinning for critical API endpoints

## Incident Response

In case of security incident:
1. Immediately revoke compromised credentials
2. Check logs for unauthorized access
3. Review backup integrity
4. Notify affected systems
5. Document incident in security log
6. Perform post-mortem

## Compliance

- Maintain audit logs for compliance (SOC2, GDPR, etc.)
- Document data retention policies
- Implement data access controls
- Regular security assessments

---

Last Updated: 2026-04-02
