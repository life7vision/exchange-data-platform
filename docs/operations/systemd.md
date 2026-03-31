# systemd Deployment

## Paths

- Repo root: `/opt/exchange-data-platform`
- Env file: `/opt/exchange-data-platform/deployments/docker/.env`
- Unit files:
  - `deployments/systemd/exchange-data-platform.service`
  - `deployments/systemd/exchange-data-platform-update.service`
  - `deployments/systemd/exchange-data-platform-update.timer`

## Install

```bash
sudo cp deployments/systemd/exchange-data-platform.service /etc/systemd/system/
sudo cp deployments/systemd/exchange-data-platform-update.service /etc/systemd/system/
sudo cp deployments/systemd/exchange-data-platform-update.timer /etc/systemd/system/
sudo cp deployments/systemd/exchange-data-platform-backup.service /etc/systemd/system/
sudo cp deployments/systemd/exchange-data-platform-backup.timer /etc/systemd/system/
sudo cp deployments/systemd/exchange-data-platform-retention.service /etc/systemd/system/
sudo cp deployments/systemd/exchange-data-platform-retention.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now exchange-data-platform.service
sudo systemctl enable --now exchange-data-platform-update.timer
sudo systemctl enable --now exchange-data-platform-backup.timer
sudo systemctl enable --now exchange-data-platform-retention.timer
```

## Operations

```bash
sudo systemctl status exchange-data-platform.service
sudo systemctl restart exchange-data-platform.service
sudo systemctl start exchange-data-platform-update.service
sudo systemctl start exchange-data-platform-backup.service
sudo systemctl start exchange-data-platform-retention.service
sudo journalctl -u exchange-data-platform.service -f
```
