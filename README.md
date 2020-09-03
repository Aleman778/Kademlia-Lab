# Kademlia-Lab
Distributed hash table for decentralized peer-to-peer computer networks

## Setup project

### Setup and run using docker on Linux
1. Clone the project
```
$ git clone https://github.com/Aleman778/kademlia-lab.git
$ cd kademlia-lab
```
2. Setup and deploy docker containers (you may need to set permissions `chmod +x kademlia_setup.sh`)
```
$ ./kademlia_setup.sh
```
3. Test connectivity by pinging another node
```
$ docker exec -it <SOURCE NODE NAME> ping <DESTINATION NODE NAME>
```

## Destroy the network
Just run the destroy script (you may need to set permissions `chmod +x kademlia_destroy.sh`)
```
$ ./kademlia_destroy.sh
```