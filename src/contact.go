package main

import (
	"fmt"
	"bytes"
	"encoding/gob"
)

// Contact definition
// stores the KademliaID, the ip address and the distance
type Contact struct {
	ID       *KademliaID
	Address  string
	distance *KademliaID
}
// NewContact returns a new instance of a Contact
func NewContact(id *KademliaID, address string) Contact {
	return Contact{id, address, nil}
}

// CalcDistance calculates the distance to the target and 
// fills the contacts distance field
func (contact *Contact) CalcDistance(target *KademliaID) {
	contact.distance = contact.ID.CalcDistance(target)
}

// String returns a simple string representation of a Contact
func (contact *Contact) String() string {
	return fmt.Sprintf(`contact("%s", "%s")`, contact.ID, contact.Address)
}

func PopCandidate(candidates []Contact) ([]Contact, Contact) {
	contact := candidates[0]
	copy(candidates, candidates[1:])
	candidates = candidates[:len(candidates)-1]
	return candidates, contact
}

func InCandidates(candidates []Contact, contact Contact) bool {
	for _, c := range candidates {
		if contact.ID.Equals(c.ID){
			return true
		}
	}
	return false
}

func EncodeContacts(contacts []Contact) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(contacts)
	checkError(err)
	return buffer.Bytes()
}

func DecodeContacts(contacts *[]Contact, data []byte) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(contacts)
	checkError(err)
}

