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
echo ____________________ CLONAR REPO
cd /home/jonbul/servers

CARPETA="ships-go"

# Comprobar si existe
if [ -d "$CARPETA" ]; then
    cd $CARPETA
    git fetch
    git pull
else
    rm -r -f $CARPETA
    git clone git@github.com:jonbul/ships-go.git
    cd $CARPETA
fi
# git checkout cleaning


cp -f /home/jonbul/servers/files/.env /home/jonbul/servers/$CARPETA/.env

echo ____________________ NUEVO DOCKER
docker build -t $IMAGE_NAME .


docker run -d -p 3000:3000 --name $CONTAINER_NAME -v /home/jonbul/servers/files/ssl:/ssl $IMAGE_NAME

docker ps -a
