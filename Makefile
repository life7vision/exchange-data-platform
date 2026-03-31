GO ?= go

.PHONY: fmt test build tidy docker-up docker-build docker-up-prod docker-down docker-logs backup restore retention

fmt:
	gofmt -w $$(find . -name '*.go' | sort)

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

build:
	$(GO) build ./...

docker-build:
	docker compose -f deployments/docker/compose.yml build

docker-up:
	docker compose -f deployments/docker/compose.yml up --build

docker-up-prod:
	docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml up -d --build

docker-down:
	docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml down

docker-logs:
	docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml logs --tail=200

backup:
	bash scripts/backup/backup_lake.sh

restore:
	bash scripts/backup/restore_lake.sh

retention:
	bash scripts/retention/purge_old_data.sh
