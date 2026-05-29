#!/bin/bash

killAndRemoveContainer() {
    if [ "$(docker ps -aq -f name=mongodb)" ]; then
        echo "Stopping and removing existing MongoDB container..."
        docker stop mongodb 2>/dev/null
        docker rm mongodb
    fi
}
runMongoCommand() {
    echo "Running MongoDB command on database $1: $2"
    docker exec -it mongodb mongosh \
      -u admin -p admin --authenticationDatabase admin \
      --eval "var db = db.getSiblingDB('$1'); $2"
}
loadTestData() {
    shopt -s nullglob
    for file in ./testData/*.json; do
        dbname=$(basename "$file" | cut -d. -f1)
        collection=$(basename "$file" | cut -d. -f2)

        echo ""
        echo "Create collection $dbname.$collection"
        runMongoCommand "$dbname" "db.createCollection('$collection');"

        echo "Loading test data from $file into $dbname.$collection"
        docker cp "$file" mongodb:/tmp/testdata.json
        docker exec mongodb mongoimport \
          -u admin -p admin --authenticationDatabase admin \
          --db "$dbname" --collection "$collection" \
          --file /tmp/testdata.json --jsonArray
    done
}

# If docker is not installed, install it

if ! command -v docker &> /dev/null
then
    echo "Docker could not be found, installing it..."

    sudo dnf install docker-cli -y
    sudo systemctl start docker

    # Add the current user to the docker group to run docker without sudo
    sudo groupadd docker
    sudo usermod -aG docker $USER

    # Set permissions for the .docker directory to allow the current user to access it
    mkdir -p /home/"$USER"/.docker
    sudo chown "$USER":"$USER" /home/"$USER"/.docker -R
    sudo chmod g+rwx "$HOME/.docker" -R

    # Docker run always at startup
    sudo systemctl enable docker

    echo "Please log out and log back in to apply docker group changes, then re-run this script."
    exit 0
fi

# Check if the mongodb container is already running, if not, run it
if [ "$(docker ps -aq -f name=mongodb)" ]; then
    echo "MongoDB container already exists (running or stopped). Do you want to recreate it? (y/n)"
    read answer
    if [ "$answer" = "y" ]; then
        killAndRemoveContainer
    else
        echo "Exiting..."
        exit 0
    fi
fi

docker run --name mongodb -d -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  mongo:7.0

echo "Waiting for MongoDB to start..."
until docker exec mongodb mongosh \
  -u admin -p admin --authenticationDatabase admin \
  --eval "db.runCommand({ ping: 1 })" &>/dev/null; do
    sleep 1
done
echo "MongoDB is up and running."

# Create user and password for mongo
echo "Creating MongoDB users..."
runMongoCommand "admin" "db.createUser({user: 'testAdmin', pwd: 'testAdmin', roles: [{ role: 'readWriteAnyDatabase', db: 'admin' },{ role: 'userAdminAnyDatabase', db: 'admin' },{ role: 'dbAdminAnyDatabase', db: 'admin' }]});"
runMongoCommand "jaes" "db.createUser({user: 'test', pwd: 'test', roles: [ { role: 'readWrite', db: 'jaes' } ]});"

# load test data from json files in ./testData/{dbname}.{collection}.json
echo "Loading test data into MongoDB..."
loadTestData

# Write exit to kill container
while true; do
    echo "Write exit to kill container: "
    read -p "" input
    if [ "$input" = "exit" ]; then
        killAndRemoveContainer
        clear
        exit 0
    fi
done