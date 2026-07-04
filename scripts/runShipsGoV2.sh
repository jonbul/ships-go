#!/bin/sh
set -x # show comands in execution

#which docker
docker ps -a
CONTAINER_NAME="ships_go-container"
IMAGE_NAME="ships_go-image"


echo "Usuario actual: $(whoami)"
echo ____________________ PARADA
docker stop $CONTAINER_NAME
echo ____________________ BORRAR DOCKER
docker rm $CONTAINER_NAME
echo ____________________ BORRAR IMAGEN DOCKER
docker rmi $IMAGE_NAME:latest

set -e # exit on error

cd /home/jonbul/servers

CARPETA="ships-go"

# Comprobar si existe
if [ ! -d "$CARPETA" ]; then
    mkdir $CARPETA
fi

#Download build file
cd $CARPETA
#ask for the latest release from github or snapshot
echo "Do you want to download the latest release or the latest snapshot?"
echo "1 release"
echo "2 snapshot"
read answer

if [ "$answer" -eq 1 ]; then
    curl -L --fail https://github.com/jonbul/ships-go/releases/latest/download/ships -o ships
elif [ "$answer" -eq 2 ]; then
    curl -L --fail https://github.com/jonbul/ships-go/releases/download/latest-snapshot/ships -o ships
fi
chmod +x ships


# Create Dockerfile
cat <<EOF > Dockerfile
FROM alpine:3.19


WORKDIR /app

# Copia el binario compilado localmente por el script
COPY ships .

# Asegura permisos de ejecución para el binario
RUN chmod +x ships

# Expone el puerto 3000
EXPOSE 3000 3001

# Ejecuta la aplicación
CMD ["./ships"]
EOF

echo ____________________ NUEVO DOCKER
docker build -t $IMAGE_NAME .

docker run -d -p 3000:3000 -p 3001:3001 --name $CONTAINER_NAME \
 -v /home/jonbul/servers/files/ssl:/ssl \
 -v /home/jonbul/servers/files/.env:/app/.env $IMAGE_NAME

docker ps -a
