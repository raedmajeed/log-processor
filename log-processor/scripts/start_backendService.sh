#!/bin/bash

cd ../docker || exit 1

docker-compose down

cd -

# Run build commands
sh ./build_MainService.sh make_all
sh ./build_MainService.sh make_docker

# Navigate to the docker directory
cd ../docker || exit 1  # Exit if the directory change fails

# Print the current working directory
echo "Current Directory: ${PWD}"


docker-compose up -d

echo "Docker Compose started successfully."
