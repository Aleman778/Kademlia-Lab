#!/bin/bash
echo "Setting up kademlia lab..."
docker build . -t kadlab
docker swarm init
docker stack deploy STACK --compose-file docker-compose.yml
echo "Kademlia lab is setup! May take some time for the containers to startup."
