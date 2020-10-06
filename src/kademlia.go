package main

import "crypto/sha1"

type Kademlia struct {
	network *Network
}

func NewKademlia(address string) *Kademlia {
	return &Kademlia{NewNetwork(address)}
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	kademlia.network.iterativeFindNode(target.ID)
}

/*
Takes a hash as its only argument, and outputs the contents of the object and the node it was retrieved from, if it could be downloaded successfully.
*/
func (kademlia *Kademlia) LookupData(hash string) ([]Contact, []byte) {
	return kademlia.network.iterativeFindData(hash)
}

/*
Takes a single argument, the contents of the file you are uploading, and outputs the hash of the object, if it could be uploaded successfully.
*/
func (kademlia *Kademlia) Store(data []byte) (string, error) {
	createHash := sha1.Sum(data)
	hash := string(createHash[:])
	hashID := NewKademliaID(hash)
	err := kademlia.network.iterativeStore(hashID, data)
	if err != nil {
		return "", err
	}
	return hash, nil
}
