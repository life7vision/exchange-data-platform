# Production Runbook

## Startup

```bash
cp deployments/docker/.env.example deployments/docker/.env
cp deployments/docker/secrets.example.env deployments/docker/secrets.env
chmod 600 deployments/docker/secrets.env
make docker-build
make docker-up-prod
```

## Health

- `GET /healthz`
- `GET /metrics`
- `GET /livez`

## Data lifecycle

1. Worker exchange API'den veri ceker.
2. Ham kayitlar kisa omurlu temp spool'a yazilir.
3. Standart Parquet dosyasi yazilir.
4. Manifest ve checkpoint kalici yazilir.
5. Temp ham dosya silinir.

## Failure behavior

- Bir exchange worker duserse sadece o exchange etkilenir.
- Reject edilen batch metadata'si `lake/rejects` altina yazilir.
- Checkpoint ve manifest sayesinde yeniden calistirma sonrasi iz surulebilir.

## Production commands

```bash
docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml ps
docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml logs --tail=200
docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml up -d --build
docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml down
./scripts/backup/backup_lake.sh
./scripts/retention/purge_old_data.sh
```

## systemd

`docs/operations/systemd.md` altindaki adimlar ile servis kalici hale getirilebilir.

## Secrets

- `deployments/docker/secrets.example.env` dosyasini `deployments/docker/secrets.env` olarak kopyalayin.
- Dosya izinlerini `600` yapin.
- Production ortaminda plain env yerine secret manager tercih edin.

## Backup and retention

- Backup script: `scripts/backup/backup_lake.sh`
- Restore script: `scripts/backup/restore_lake.sh`
- Retention purge: `scripts/retention/purge_old_data.sh`

Varsayilan retention degerleri `.env` icinden yonetilir.
