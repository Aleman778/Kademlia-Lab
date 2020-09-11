#!/bin/bash
echo "Destroying up kademlia lab..."
docker stack rm STACK
docker swarm leave --force
echo "Kademlia lab is destroyed!"
