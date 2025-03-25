#!/bin/bash

cd ../docker || exit 1

docker-compose down

cd -

sh ./build_MainService.sh make_all
sh ./build_MainService.sh make_docker

cd ../docker || exit 1

echo "Current Directory: ${PWD}"

docker-compose up -d

echo "Docker Compose started successfully."
