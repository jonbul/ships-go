#!/bin/sh
set -x # show commands in execution

CONTAINER_NAME="ships-go"
IMAGE_NAME="ships-go-image"
ROOT_PATH="/home/jonbul/servers"
PROJECT_PATH="$ROOT_PATH/ships-go"
SSL_PATH="$ROOT_PATH/files/ssl"
ENV_PATH="$ROOT_PATH/files/.env"


echo "Usuario actual: $(whoami)"

echo "=== PARADA ==="
docker stop $CONTAINER_NAME 2>/dev/null || true
echo "=== BORRAR CONTENEDOR ==="
docker rm $CONTAINER_NAME 2>/dev/null || true
echo "=== BORRAR IMAGEN ==="
docker rmi $IMAGE_NAME:latest 2>/dev/null || true

set -e # exit on error

# Crear directorio de trabajo si no existe
mkdir -p "$PROJECT_PATH"
cd "$PROJECT_PATH"

# Descargar el binario compilado por GitHub Actions
echo "¿Descargar la última release o el último snapshot?"
echo "1) release"
echo "2) snapshot"
read -r answer

if [ "$answer" -eq 1 ]; then
    curl -L --fail https://github.com/jonbul/ships-go/releases/latest/download/ships -o ships
elif [ "$answer" -eq 2 ]; then
    curl -L --fail https://github.com/jonbul/ships-go/releases/download/latest-snapshot/ships -o ships
else
    echo "Opción inválida"
    exit 1
fi
chmod +x ships

echo "=== CONSTRUIR IMAGEN DOCKER ==="
# Se genera el Dockerfile en tiempo de ejecución, sin depender de ningún archivo externo
cat > Dockerfile.tmp << 'DOCKERFILE'
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY ships .
RUN chmod +x ships
EXPOSE 3000 3001
CMD ["./ships"]
DOCKERFILE

docker build -f Dockerfile.tmp -t $IMAGE_NAME:latest .
rm Dockerfile.tmp

echo "=== CREAR RED DOCKER (si no existe) ==="
docker network inspect ships-network >/dev/null 2>&1 || docker network create ships-network

echo "=== ARRANCAR CONTENEDOR ==="
# Monta el fichero .env desde el directorio de trabajo (debe existir previamente)
# Los certificados SSL referenciados en .env también deben estar accesibles desde el contenedor
docker run -d \
    --name $CONTAINER_NAME \
    --network ships-network \
    -p 3000:3000 \
    -p 3001:3001 \
    -v "$ENV_PATH:/app/.env:ro" \
    -v "$SSL_PATH:/ssl:ro" \
    $IMAGE_NAME:latest

docker ps -a
