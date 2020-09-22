package main

import (
	"fmt"
	"net"
	"os"
	"crypto/sha1"
	"encoding/gob"
)

const k = 5
const ALPHA = 3

type Network struct {
	table *RoutingTable
}

type Rpc int

const (
	findData Rpc = iota
	sendData
)

type MSG struct {
	Msg Rpc
	Data []byte
	Me Contact
}

func (network *Network) Listen(port string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	checkError(err)

	conn, err := net.ListenUDP("udp4", udpAddr)
	checkError(err)
	defer conn.Close()

	fmt.Println("Sever setup finished")

	for {
		network.HandleClient(conn)
	}
}

func (network *Network) SendPingMessage(contact *Contact) bool {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: nil}

	conn := rpcMsg.SendTo(contact.Address)
	defer conn.Close()

	_, _, err := GetRPCMessage(conn, 15)
	if err != nil {
		fmt.Println("\nGeting response timeout\n")
		return false
	}

	return true
}

func (network *Network) InitNodeLookup(address string) {
	var called ContactCandidates
	me := network.table.GetMe()
	me.CalcDistance(me.ID)
	called.Add(me)

	var notCalled ContactCandidates
	contacts, err := network.SendFindContactMessage(address, *me.ID)
	checkError(err)

	for _, contact := range contacts {
		if !notCalled.InCandidates(contact) && !called.InCandidates(contact) {
			contact.CalcDistance(me.ID)
			notCalled.Add(contact)
		}
	}

	notCalled.Sort()

	network.HandleNodeLookupGoRoutines(*me.ID, called, notCalled)
}

func (network *Network) NodeLookup(id KademliaID) []Contact {
	var called ContactCandidates
	me := network.table.GetMe()
	me.CalcDistance(&id)
	called.Add(me)

	var notCalled ContactCandidates
	notCalled.Append(network.table.FindClosestContacts(&id, k))
	notCalled.Sort()

	return network.HandleNodeLookupGoRoutines(id, called, notCalled)
}

func (network *Network) HandleNodeLookupGoRoutines(id KademliaID, called ContactCandidates, notCalled ContactCandidates) []Contact {
	readCh := make(chan []Contact)
	defer close(readCh)

	writeCh := make(chan Contact)
	defer close(writeCh)

	numNotRunning := ALPHA
	isDone := false

	// Start ALPHA GoRoutines
	for i := 0; i < ALPHA; i++ {
		go network.NodeLookupGoRoutine(id, readCh, writeCh)
		if notCalled.Len() > 0 {
			numNotRunning -= 1
			contact := notCalled.Drop(0)
			called.Add(contact)
			writeCh <- contact
		}
	}
	called.Sort()


	for {
		if numNotRunning == ALPHA {
			length := called.Len()
			if length < k {
				return called.GetContacts(length)
			}
			return called.GetContacts(k)
		}

		// Get contacts from one of the GoRoutines respones messages
		contacts := <-readCh
		numNotRunning += 1

		// Add uncontacted nodes to the notCalled list.
		for _, contact := range contacts {
			if !notCalled.InCandidates(contact) && !called.InCandidates(contact) {
				contact.CalcDistance(&id)
				notCalled.Add(contact)
			}
		}
		notCalled.Sort()


		// Make all not running GoRoutines send findnode rpc if there are any uncontacted contacts
		for {
			if notCalled.Len() != 0 && called.Len() >= k  {
				isDone = called.Get(k-1).distance.Less(notCalled.Get(0).distance)
			}

			if numNotRunning < 1 || notCalled.Len() == 0 || isDone {
				break
			}
			numNotRunning -= 1
			contact := notCalled.Drop(0)
			called.Add(contact)
			writeCh <- contact
		}
		called.Sort()
	}
}

// GoRoutine that sends FindNode
func (network *Network) NodeLookupGoRoutine(id KademliaID, writeCh chan<- []Contact, readCh <-chan Contact) {
	for {
		contact, more := <-readCh
		if !more {
			return
		}

		contacts, err := network.SendFindContactMessage(contact.Address, id)
		if err != nil {
			writeCh <- []Contact{}
		} else {
			writeCh <- contacts
		}
	}
}


func (network *Network) SendFindDataMessage(hash string) {
	hashID := NewKademliaID(hash)
	closest := network.table.FindClosestContacts(hashID, k)

	for _, contact := range closest {
		go func(address string) {
			conn, err := net.Dial("udp4", address)
			defer conn.Close()
			if err != nil {
				fmt.Errorf("Error in SendFindDataMessage: %v", err)
			} else {
				msg := []byte("FINDD\n")
				conn.Write(msg)
			}
		}(contact.Address)
	}
}

func (network *Network) SendStoreMessage(data []byte) {
	hash := sha1.Sum(data)
	hashID := NewKademliaID(string(hash[:]))
	closest := network.table.FindClosestContacts(hashID, k)

	for _, contact := range closest {
		go func(address string) {
			conn, err := net.Dial("udp4", address) //This might go outside go func
			defer conn.Close()
			if err != nil {
				fmt.Errorf("Error in net.Dial: %v", err)
			} else {
				msg := MSG{sendData, data, network.table.GetMe()}
				enc := gob.NewEncoder(conn)
				err := enc.Encode(msg)
				if err != nil {
					fmt.Errorf("Error in enc.Encode: %v", err)
				}
			}
		}(contact.Address)
	}
}



func (network *Network) SendFindContactMessage(addr string, id KademliaID) ([]Contact, error) {
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: EncodeKademliaID(id)}

	conn := rpcMsg.SendTo(addr)
	defer conn.Close()

	responesMsg, _, err := GetRPCMessage(conn, 15)
	if err != nil {
		fmt.Println("\nGeting response timeout")
		return []Contact{}, err
	}

	if responesMsg.IsNode {
		network.AddContact(responesMsg.Sender)
	}

	var contacts []Contact
	DecodeContacts(&contacts, responesMsg.Data)
	return contacts, nil
}

func (network *Network) HandleClient(conn *net.UDPConn) {
	rpcMsg, addr, err := GetRPCMessage(conn, 0)

	if err != nil {
		return
	}

	if rpcMsg.IsNode {
		network.AddContact(rpcMsg.Sender)
	}

	switch rpcMsg.Type {
	case Ping:
		network.HandlePingMessage(rpcMsg.Data, conn, addr)
	case Store:
		network.HandleStoreMessage(rpcMsg.Data, conn, addr)
	case FindNode:
		network.HandleFindNodeMessage(rpcMsg.Data, conn, addr)
	case FindValue:
		network.HandleFindValueMessage(rpcMsg.Data, conn, addr)
	case ExitNode:
		network.HandleExitNodeMessage(conn, addr)
	case Test:
		var id KademliaID
		DecodeKademliaID(&id, rpcMsg.Data)
		responseMsg := RPCMessage{
			Type: Test,
			IsNode: true,
			Sender: network.table.GetMe(),
			Data: EncodeContacts(network.NodeLookup(id))}
		responseMsg.SendResponse(conn, addr)
	}
}

func (network *Network) HandlePingMessage(Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: Data}
	rpcMsg.SendResponse(conn, addr)
}

func (network *Network) HandleStoreMessage(Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Store,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: Data}
	rpcMsg.SendResponse(conn, addr)
	//TODO
}

func (network *Network) HandleFindNodeMessage(Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	var id KademliaID
	DecodeKademliaID(&id, Data)
	contacts := network.table.FindClosestContacts(&id, k)
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: EncodeContacts(contacts)}
	rpcMsg.SendResponse(conn, addr)
}

func (network *Network) HandleFindValueMessage(Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: FindValue,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: Data}
	rpcMsg.SendResponse(conn, addr)
	//TODO
}

func (network *Network) HandleExitNodeMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: ExitNode,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: nil}
	rpcMsg.SendResponse(conn, addr)

	fmt.Println("Shutting down server")
	os.Exit(0);
}


func (network *Network) AddContact(contact Contact) {
	bucketIndex := network.table.getBucketIndex(contact.ID)

	if network.table.IsBucketFull(bucketIndex) {
		lastContact := network.table.GetLastInBucket(bucketIndex)
		isAlive := network.SendPingMessage(&lastContact)
		if isAlive {
			return
		}

		network.table.RemoveContactFromBucket(bucketIndex, lastContact)
		network.table.AddContact(contact)
	} else {
		network.table.AddContact(contact)
	}
}

