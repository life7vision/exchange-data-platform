#!/usr/bin/env bash
set -euo pipefail

docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml ps
echo
docker compose -f deployments/docker/compose.yml -f deployments/docker/compose.prod.yml logs --tail=50
