#!/bin/sh
set -x # show commands in execution

ROOT_PATH="/home/jonbul/servers/monitoring"

echo "=== CREAR RED DOCKER (si no existe) ==="
docker network inspect ships-network >/dev/null 2>&1 || docker network create ships-network

echo "=== CREAR DIRECTORIOS DE DATOS (si no existen) ==="
mkdir -p "$ROOT_PATH/prometheus_data"
mkdir -p "$ROOT_PATH/grafana_data"

# Permisos requeridos por cada imagen
sudo chown -R 65534:65534 "$ROOT_PATH/prometheus_data"
sudo chown -R 472:472     "$ROOT_PATH/grafana_data"

echo "=== ARRANCAR CONTENEDORES ==="
docker compose -f "$ROOT_PATH/docker-compose.yml" up -d

docker ps -a