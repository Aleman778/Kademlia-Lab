package main

import (
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
)

type Network struct {
	table   *RoutingTable
	storage *Storage
}

const k = 5
const a = 3

func NewNetwork(address string) *Network {
	return &Network{NewRoutingTable(NewContact(NewRandomKademliaID(), address)), NewStorage()}
}

func (network *Network) Listen(ip string, port int) {
	addr, err := net.ResolveUDPAddr("udp", string(port))
	if err != nil {
		fmt.Errorf("Could not use port: ", port)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Errorf("Could not listen")
		return
	}

	defer conn.Close()

	for {
		var recv Payload
		decoder := gob.NewDecoder(conn)
		err := decoder.Decode(recv)

		if err != nil {
			fmt.Errorf("Could not decode msg")
			continue
		}

		switch recv.Msg {
		case PING:
			network.SendPingMessage(recv.Me)
		case STORE:
			network.storage.Store(recv.Hash, recv.Data)
		case FINDDATA:
			data, ok := network.storage.Load(recv.Hash)
			if ok {
				sendPayload(recv.Me.Address, &Payload{FINDDATA, recv.Hash, data, nil, &network.table.me})
			} else {
				hashID := NewKademliaID(recv.Hash)
				closetContacts := network.table.FindClosestContacts(hashID, k)
				sendPayload(recv.Me.Address, &Payload{FINDDATA, "", nil, closetContacts, &network.table.me})
			}
		case FINDNODE:
			contact := recv.Contacts[0]
			closetContacts := network.table.FindClosestContacts(contact.ID, k)
			sendPayload(recv.Me.Address, &Payload{FINDDATA, "", nil, closetContacts, &network.table.me})
		}
	}
}

/*
Comments taken from http://xlattice.sourceforge.net/components/protocol/kademlia/specs.html#protocol
*/

/*
This RPC involves one node sending a PING message to another, which presumably replies with a PONG.
This has a two-fold effect: the recipient of the PING must update the bucket corresponding to the sender;
and, if there is a reply, the sender must update the bucket appropriate to the recipient.
All RPC packets are required to carry an RPC identifier assigned by the sender and echoed in the reply. This is a quasi-random number of length B (160 bits).
*/
func (network *Network) SendPingMessage(contact *Contact) {
	sendPayload(contact.Address, &Payload{PING, "", nil, nil, &network.table.me})
}

/*
The FIND_NODE RPC includes a 160-bit key. The recipient of the RPC returns up to k triples (IP address, port, nodeID)
for the contacts that it knows to be closest to the key.
The recipient must return k triples if at all possible. It may only return fewer than k if it is returning all of the contacts that it has knowledge of.
This is a primitive operation, not an iterative one.
*/
func (network *Network) SendFindContactMessage(contact *Contact) []Contact {
	contacts := make([]Contact, 1)
	contacts[0] = *contact
	recv, _ := sendAndRecvPayload(contact.Address, &Payload{FINDNODE, "", nil, contacts, &network.table.me})
	return recv.Contacts
}

/*
A FIND_VALUE RPC includes a B=160-bit key. If a corresponding value is present on the recipient, the associated data is returned.
Otherwise the RPC is equivalent to a FIND_NODE and a set of k triples is returned.
This is a primitive operation, not an iterative one.
*/
func (network *Network) SendFindDataMessage(hash string, contact *Contact) ([]Contact, []byte) {
	recv, _ := sendAndRecvPayload(contact.Address, &Payload{FINDDATA, hash, nil, nil, &network.table.me})
	if recv.Hash == hash {
		return nil, recv.Data
	}
	return recv.Contacts, nil
}

/*
The sender of the STORE RPC provides a key and a block of data and requires that the recipient store the data and make it available for later retrieval by that key.
This is a primitive operation, not an iterative one.
*/
func (network *Network) SendStoreMessage(data []byte, contact *Contact) error {
	createHash := sha1.Sum(data)
	hash := string(createHash[:])
	_, err := sendAndRecvPayload(contact.Address, &Payload{STORE, hash, data, nil, &network.table.me})
	return err
}

/*
The search begins by selecting alpha contacts from the non-empty k-bucket closest to the bucket appropriate to the key being searched on.
If there are fewer than alpha contacts in that bucket, contacts are selected from other buckets. The contact closest to the target key, closestNode, is noted.
The first alpha contacts selected are used to create a shortlist for the search.
The node then sends parallel, asynchronous FIND_* RPCs to the alpha contacts in the shortlist.
Each contact, if it is live, should normally return k triples. If any of the alpha contacts fails to reply, it is removed from the shortlist, at least temporarily.
The node then fills the shortlist with contacts from the replies received. These are those closest to the target. From the shortlist it selects another alpha contacts. The only condition for this selection is that they have not already been contacted. Once again a FIND_* RPC is sent to each in parallel.
Each such parallel search updates closestNode, the closest node seen so far.
The sequence of parallel searches is continued until either no node in the sets returned is closer
than the closest node already seen or the initiating node has accumulated k probed and known to be active contacts.
If a cycle doesn't find a closer node, if closestNode is unchanged, then the initiating node sends a FIND_* RPC to each of the k closest nodes
that it has not already queried.
At the end of this process, the node will have accumulated a set of k active contacts or (if the RPC was FIND_VALUE) may have found a data value.
Either a set of triples or the value is returned to the caller.
*/
/*func (network *Network) iterativeNodeLookup(contact *Contact) { //TODO: parallel FIND_NODE calls
	closetContacts := network.table.closetContacts(contact, a)
	closestNode := closetContacts[0]
	for {
		var shortlist ContactCandidates
		for _, contact := range closetContacts {
			recvContacts := SendFindContactMessage(&contact)
			shortlist.append(recvContacts)
		}
		shortlist.Sort()
		closetContacts := shortlist.GetContacts(a)
		if closetNode == closetContacts[0] {
			break
		}
		closestNode := closetContacts[0]
	}
}*/

/*
This is the Kademlia store operation. The initiating node does an iterativeFindNode,
collecting a set of k closest contacts, and then sends a primitive STORE RPC to each.
iterativeStores are used for publishing or replicating data on a Kademlia network.
*/
func (network *Network) iterativeStore(id *KademliaID, data []byte) error {
	contacts := network.iterativeFindNode(id)
	good := false
	for _, contact := range contacts {
		err := network.SendStoreMessage(data, &contact)
		if err == nil {
			good = true
		}
	}
	if good {
		return nil
	}
	return errors.New("Could not store data")
}

/*
This is the basic Kademlia node lookup operation.
As described above, the initiating node builds a list of k "closest" contacts using iterative node lookup and the FIND_NODE RPC.
The list is returned to the caller.
*/
func (network *Network) iterativeFindNode(id *KademliaID) []Contact {
	closetContacts := network.table.FindClosestContacts(id, a)
	closestNode := closetContacts[0]
	for {
		var shortlist ContactCandidates
		for _, contact := range closetContacts {
			recvContacts := network.SendFindContactMessage(&contact)
			shortlist.Append(recvContacts)
		}
		shortlist.Sort()
		closetContacts := shortlist.GetContacts(a)
		if closestNode == closetContacts[0] {
			return shortlist.GetContacts(k)
		}
		closestNode = closetContacts[0]
	}
}

/*
This is the Kademlia search operation.
It is conducted as a node lookup, and so builds a list of k closest contacts.
However, this is done using the FIND_VALUE RPC instead of the FIND_NODE RPC.
If at any time during the node lookup the value is returned instead of a set of contacts, the search is abandoned and the value is returned.
Otherwise, if no value has been found, the list of k closest contacts is returned to the caller.
When an iterativeFindValue succeeds, the initiator must store the key/value pair at the closest node seen which did not return the value.
*/
func (network *Network) iterativeFindData(hash string) ([]Contact, []byte) {
	hashID := NewKademliaID(hash)
	closetContacts := network.table.FindClosestContacts(hashID, a)
	closestNode := closetContacts[0]
	for {
		var shortlist ContactCandidates
		for _, contact := range closetContacts {
			recvContacts, recvData := network.SendFindDataMessage(hash, &contact)
			if recvData != nil {
				contacts := make([]Contact, 1)
				contacts[0] = contact
				return contacts, recvData
			}
			shortlist.Append(recvContacts)
		}
		shortlist.Sort()
		closetContacts := shortlist.GetContacts(a)
		if closestNode == closetContacts[0] {
			return shortlist.GetContacts(k), nil
		}
		closestNode = closetContacts[0]
	}
}
