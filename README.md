# Exchange Data Platform

Go tabanli, moduler, exchange-basina ayri servislerle calisan veri toplama platformu.

## Ozellikler

- `binance`, `bybit`, `binance_tr`, `bybit_tr` icin ayri worker servisleri
- temp raw spool -> Parquet -> manifest/checkpoint akisi
- Parquet basariyla yazildiginda temp ham dosyalarin silinmesi
- Docker Compose ile servis izolasyonu
- ortak connector contract ve exchange-bazli moduller

## Hizli Baslangic

```bash
go test ./...
go build ./...
docker compose -f deployments/docker/compose.yml up --build
```

## Production

- Base Compose: `deployments/docker/compose.yml`
- Production override: `deployments/docker/compose.prod.yml`
- Env template: `deployments/docker/.env.example`
- Secrets template: `deployments/docker/secrets.example.env`
- systemd units: `deployments/systemd/`

Kalici servis kurulumu ve operasyon adimlari icin `docs/operations/runbook.md` ve `docs/operations/systemd.md` kullanilir.
