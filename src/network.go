package main

import (
	"fmt"
    "time"
	"net"
	"os"
    "sort"
	"crypto/sha1"
)

const k = 5
const ALPHA = 3

type Network struct {
	table *RoutingTable
	storage *Storage
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

func GetRPCMessage(conn *net.UDPConn, timeout time.Duration) (RPCMessage, *net.UDPAddr, error) {
	var rpcMsg RPCMessage
	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout * time.Second))
	}
	inputBytes := make([]byte, 1024)
	length, addr, err := conn.ReadFromUDP(inputBytes)
	if err != nil {
		return rpcMsg, nil, err
	}

	DecodeRPCMessage(&rpcMsg, inputBytes[:length])
	fmt.Println("Recived Msg from ", addr, " :\n", rpcMsg.String())

	return rpcMsg, addr, nil
}

func (network *Network) SendPingMessage(contact *Contact) bool {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{"", nil, nil}}

	conn := rpcMsg.SendTo(contact.Address)
	defer conn.Close()

	_, _, err := GetRPCMessage(conn, 15)
	if err != nil {
		fmt.Println("\nGeting response timeout")
		return false
	}

	return true
}

func (network *Network) BootstrapNode(address string) {
	var called []Contact
	me := network.table.GetMe()
	me.CalcDistance(me.ID)
	called = append(called, me)

	var notCalled []Contact
	contacts, err := network.SendFindContactMessage(address, *me.ID)
	checkError(err)

	for _, contact := range contacts {
		if !InCandidates(notCalled, contact) && !InCandidates(called, contact) {
			contact.CalcDistance(me.ID)
			notCalled = append(notCalled, contact)
		}
	}

	sort.Sort(ByDistance(notCalled))
	network.HandleNodeLookupGoRoutines(*me.ID, called, notCalled)
}

func (network *Network) NodeLookup(id KademliaID) []Contact {
	var called []Contact
	me := network.table.GetMe()
	me.CalcDistance(&id)
	called = append(called, me)

	var notCalled []Contact
	notCalled = append(notCalled, network.table.FindClosestContacts(&id, k)...)
	sort.Sort(ByDistance(notCalled))

	return network.HandleNodeLookupGoRoutines(id, called, notCalled)
}

func (network *Network) HandleNodeLookupGoRoutines(id KademliaID, called []Contact, notCalled []Contact) []Contact {
	readCh := make(chan []Contact)
	defer close(readCh)

	writeCh := make(chan Contact)
	defer close(writeCh)

	numNotRunning := ALPHA
	isDone := false

	// Start ALPHA GoRoutines
	for i := 0; i < ALPHA; i++ {
		go network.NodeLookupGoRoutine(id, readCh, writeCh)
		if len(notCalled) > 0 {
			numNotRunning -= 1
            var contact Contact
			notCalled, contact = PopCandidate(notCalled)
			called = append(called, contact)
			writeCh <- contact
		}
	}
    sort.Sort(ByDistance(called))

	for {
		if numNotRunning == ALPHA {
			length := len(called)
			if length < k {
				return called
			}
			return called[:k]
		}

		// Get contacts from one of the GoRoutines respones messages
		contacts := <-readCh
		numNotRunning += 1

		// Add uncontacted nodes to the notCalled list.
		for _, contact := range contacts {
			if !InCandidates(notCalled, contact) && !InCandidates(called, contact) {
				contact.CalcDistance(&id)
				notCalled = append(notCalled, contact)
			}
		}
        sort.Sort(ByDistance(notCalled))

		// Make all not running GoRoutines send findnode rpc if there are any uncontacted contacts
		for {
			if len(notCalled) != 0 && len(called) >= k  {
				isDone = called[k-1].distance.Less(notCalled[0].distance)
			}

			if numNotRunning < 1 || len(notCalled) == 0 || isDone {
				break
			}
			numNotRunning -= 1
            var contact Contact
			notCalled, contact = PopCandidate(notCalled)
			called = append(called, contact)
			writeCh <- contact
		}
        sort.Sort(ByDistance(called))
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
			rpcMsg := RPCMessage{
				Type: FindValue,
				IsNode: true,
				Sender: network.table.GetMe(),
				Payload: Payload {
					Hash: hash,
					Data: nil,
					Contacts: nil,
				}}
			conn := rpcMsg.SendTo(address)
			defer conn.Close()
		}(contact.Address)
	}
}

func (network *Network) SendStoreMessage(data []byte) {
	hash := sha1.Sum(data)
	hashID := NewKademliaID(string(hash[:]))
	closest := network.table.FindClosestContacts(hashID, k)

	for _, contact := range closest {
		go func(address string) {
			rpcMsg := RPCMessage{
				Type: Store,
				IsNode: true,
				Sender: network.table.GetMe(),
				Payload: Payload {
					Hash: string(hash[:]),
					Data: data,
					Contacts: nil,
				}}
			conn := rpcMsg.SendTo(address)
			defer conn.Close()
		}(contact.Address)
	}
}

func (network *Network) SendFindContactMessage(addr string, id KademliaID) ([]Contact, error) {
	contacts := make([]Contact, 1)
	contacts[0] = NewContact(&id, "")
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{
			Hash: "",
			Data: nil,
			Contacts: contacts,
		}}

	conn := rpcMsg.SendTo(addr)
	defer conn.Close()

	responseMsg, _, err := GetRPCMessage(conn, 15)
	if err != nil {
		fmt.Println("\nGeting response timeout")
		return []Contact{}, err
	}

	if responseMsg.IsNode {
		network.AddContact(responseMsg.Sender)
	}

	return responseMsg.Payload.Contacts, nil
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
		network.HandlePingMessage(conn, addr)
	case Store:
		network.HandleStoreMessage(&rpcMsg, conn, addr)
	case FindNode:
		network.HandleFindNodeMessage(&rpcMsg, conn, addr)
	case FindValue:
		network.HandleFindValueMessage(&rpcMsg, conn, addr)
	case ExitNode:
		network.HandleExitNodeMessage(conn, addr)
	case CliPut:
		network.HandleCliPutMessage(conn, addr)
	case CliGet:
		network.HandleCliGetMessage(conn, addr)
	case CliExit:
		network.HandleCliExitMessage(conn, addr)
	case Test:
		responseMsg := RPCMessage{
			Type: Test,
			IsNode: true,
			Sender: network.table.GetMe(),
			Payload: Payload{"", nil, rpcMsg.Payload.Contacts}}
		responseMsg.SendResponse(conn, addr)
	}
}

func (network *Network) HandlePingMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{"",nil,nil,}}
	rpcMsg.SendResponse(conn, addr)
}

func (network *Network) HandleStoreMessage(msg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Store,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{"", nil, nil}}
	rpcMsg.SendResponse(conn, addr)

	network.storage.Store(msg.Payload.Hash, msg.Payload.Data)
}

func (network *Network) HandleFindNodeMessage(msg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	var id KademliaID = *msg.Payload.Contacts[0].ID
	contacts := network.table.FindClosestContacts(&id, k)
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{
			Hash: "",
			Data: nil,
			Contacts: contacts,
		}}
	rpcMsg.SendResponse(conn, addr)
}

func (network *Network) HandleFindValueMessage(msg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	if val, ok := network.storage.Load(msg.Payload.Hash); ok {
		rpcMsg := RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: network.table.GetMe(),
			Payload: Payload{
				Hash: msg.Payload.Hash,
				Data: val,
				Contacts: nil,
			}}
			rpcMsg.SendResponse(conn, addr)
	} else {
		hashID := NewKademliaID(msg.Payload.Hash)
		closest := network.table.FindClosestContacts(hashID, k)
		rpcMsg := RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: network.table.GetMe(),
			Payload: Payload{
				Hash: "",
				Data: nil,
				Contacts: closest,
			}}
		rpcMsg.SendResponse(conn, addr)
	}
}

func (network *Network) HandleCliPutMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: CliPut,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: nil}
    // TODO: put specific value from the distributed hash table
	rpcMsg.SendResponse(conn, addr)
}

func (network *Network) HandleCliGetMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: CliGet,
		IsNode: true,
		Sender: network.table.GetMe(),
		Data: nil}
    // TODO: get the value given a hash from the distributed hash table
	rpcMsg.SendResponse(conn, addr)
    
}

func (network *Network) HandleCliExitMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: CliExit,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{"", nil, nil}}
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

