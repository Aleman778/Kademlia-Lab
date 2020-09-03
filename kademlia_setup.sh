#!/bin/bash
echo "Setting up kademlia lab..."
docker build . -t kadlab
docker swarm init
docker stack deploy STACK --compose-file docker-compose.yml
echo "Kademlia lab is setup! Check container status using `docker ps`"
