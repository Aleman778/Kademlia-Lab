# Kademlia-Lab
Distributed hash table for decentralized peer-to-peer computer networks

## Setup project

### Setup and run using docker on Linux
1. Clone the project
```
$ git clone https://github.com/Aleman778/kademlia-lab.git
$ cd kademlia-lab
```
2. Setup and deploy docker containers
```
$ ./kademlia_setup
```
3. Test connectivity by pinging another node
```
$ docker exec -it <SOURCE NODE NAME> ping <DESTINATION NODE NAME>
```

## Destroy the network
Just run the destroy script
```
$ ./kademlia_destroy
```