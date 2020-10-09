package main

import (
	"fmt"
    "time"
	"net"
	"os"
    "sort"
	"sync"
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

	fmt.Println("Server setup finished")

	for {
		network.HandleClient(conn)
	}
}

func GetRPCMessage(conn *net.UDPConn, timeout time.Duration, verbose bool) (RPCMessage, *net.UDPAddr, error) {
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
    if verbose {
        fmt.Println("Recived Msg from ", addr, " :\n", rpcMsg.String())
    }

	return rpcMsg, addr, nil
}

func (network *Network) SendPingMessage(contact *Contact) bool {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{"", nil, nil}}

	conn := rpcMsg.SendTo(contact.Address, true)
	defer conn.Close()

	_, _, err := GetRPCMessage(conn, 5, true)
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
	network.StartNodeLookup(*me.ID, notCalled)
}

func (network *Network) NodeLookup(id KademliaID) []Contact {
	var notCalled []Contact
	notCalled = append(notCalled, network.table.FindClosestContacts(&id, k)...)
	sort.Sort(ByDistance(notCalled))

	return network.StartNodeLookup(id, notCalled)
}

func (network *Network) StartNodeLookup(id KademliaID, notCalled []Contact) []Contact {
	contactsCh := make(chan LookupResponse)
	defer close(contactsCh)

	contactCh := make(chan Contact)
	defer close(contactCh)

	go network.NodeLookupSender(id, contactsCh, contactCh)

	contacts, _ := RunLookup(&id, network.table.GetMe(), notCalled, contactCh, contactsCh)
	return contacts
}



func (network *Network) NodeLookupSender(id KademliaID, writeCh chan<- LookupResponse, readCh <-chan Contact) {
	for {
		contact, more := <-readCh
		if !more {
			return
		}
		go func(writeCh chan<- LookupResponse, contact Contact) {
			contacts, err := network.SendFindContactMessage(contact.Address, id)
			if err != nil {
				writeCh <- LookupResponse{[]Contact{}, contact, false}
			} else {
				writeCh <- LookupResponse{contacts, contact, false}
			}
		}(writeCh, contact)
	}
}

func (network *Network) ValueLookup(hash string) Payload {
	if val, ok := network.storage.Load(hash); ok {
		return Payload{
			Hash: hash,
			Data: val,
			Contacts: nil};
	}

	id := NewHashedID(hash)
	var called []Contact
	me := network.table.GetMe()
	me.CalcDistance(&id)
	called = append(called, me)

	var notCalled []Contact
	notCalled = append(notCalled, network.table.FindClosestContacts(&id, k)...)
	sort.Sort(ByDistance(notCalled))

	resultCh := make(chan Payload)

	intermediateCh := make(chan Payload)
	inbetweenCh := make(chan Payload)
	go func() {
		defer close(inbetweenCh)

		payload := <-intermediateCh
		close(intermediateCh)
		resultCh <- payload
		close(resultCh)
		inbetweenCh <- payload
	}()

	go func(){

		contactsCh := make(chan LookupResponse)
		defer close(contactsCh)

		contactCh := make(chan Contact)
		defer close(contactCh)

		go network.ValueLookupSender(hash, contactsCh, contactCh, intermediateCh)

		_, contact := RunLookup(&id, network.table.GetMe(), notCalled, contactCh, contactsCh)

		payload := <-inbetweenCh
		go func(address string) {
			rpcMsg := RPCMessage{
				Type: Store,
				IsNode: true,
				Sender: network.table.GetMe(),
				Payload: payload}
			conn := rpcMsg.SendTo(address, true)
			defer conn.Close()
		}(contact.Address)
	}()

	payload := <-resultCh
	return payload
}

func (network *Network) ValueLookupSender(hash string, writeCh chan<- LookupResponse, readCh <-chan Contact, resultCh chan<- Payload) {
	isDone := false
	mutex := sync.RWMutex{}
	for {
		contact, more := <-readCh
		if !more {
			if !isDone {
				mutex.Lock()
				isDone = true
				mutex.Unlock()

				resultCh <- Payload{"", nil, nil}
			}
			return
		}
		go func(writeCh chan<- LookupResponse, contact Contact) {
			payload, err := network.SendFindDataMessage(contact.Address, hash)
			if err != nil {
				writeCh <- LookupResponse{[]Contact{}, contact, true}
				return
			}

			if payload.Data != nil && !isDone {
				resultCh <- payload
				mutex.Lock()
				isDone = true
				mutex.Unlock()
			}

			hasValue := payload.Data != nil
			writeCh <- LookupResponse{payload.Contacts, contact, hasValue}
		}(writeCh, contact)
	}
}

func (network *Network) SendFindDataMessage(address string, hash string) (Payload, error) {
	rpcMsg := RPCMessage{
		Type: FindValue,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload {
			Hash: hash,
			Data: nil,
			Contacts: nil,
		}}
	conn := rpcMsg.SendTo(address, true)
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, 15, true)
	if err != nil {
		fmt.Println("\nGetting response timeout")
		return response.Payload, err
	}
	return response.Payload, nil
}

/* TODO: unused function
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
			conn := rpcMsg.SendTo(address, true)
			defer conn.Close()
		}(contact.Address)
	}
}
*/

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

	conn := rpcMsg.SendTo(addr, true)
	defer conn.Close()

	responseMsg, _, err := GetRPCMessage(conn, 15, true)
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
	rpcMsg, addr, err := GetRPCMessage(conn, 0, true)

	if err != nil {
		return
	}

	if rpcMsg.IsNode {
		network.AddContact(rpcMsg.Sender)
	}

	switch rpcMsg.Type {
	case Ping:
		go network.HandlePingMessage(conn, addr)
	case Store:
		go network.HandleStoreMessage(&rpcMsg, conn, addr)
	case FindNode:
		go network.HandleFindNodeMessage(&rpcMsg, conn, addr)
	case FindValue:
		go network.HandleFindValueMessage(&rpcMsg, conn, addr)
	case CliPut:
		go network.HandleCliPutMessage(&rpcMsg, conn, addr)
	case CliGet:
		go network.HandleCliGetMessage(&rpcMsg,conn, addr)
	case CliExit:
		network.HandleCliExitMessage(conn, addr)
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
	network.storage.Store(msg.Payload.Hash, msg.Payload.Data)
	rpcMsg.SendResponse(conn, addr)
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

func (network *Network) HandleCliPutMessage(rpcMsg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	id := NewHashedID(rpcMsg.Payload.Hash)
	closest := network.NodeLookup(id)
	for _, c := range closest {
		go func(address string) {
			rpcMsg := RPCMessage{
				Type: Store,
				IsNode: true,
				Sender: network.table.GetMe(),
				Payload: Payload {
					Hash: string(rpcMsg.Payload.Hash),
					Data: rpcMsg.Payload.Data,
					Contacts: nil,
				}}
			conn := rpcMsg.SendTo(address, true)
			defer conn.Close()
		}(c.Address)
	}
	response := RPCMessage{
		Type: CliPut,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: Payload{rpcMsg.Payload.Hash, nil, nil}}
	response.SendResponse(conn, addr)
}

func (network *Network) HandleCliGetMessage(rpcMsg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	payload := network.ValueLookup(rpcMsg.Payload.Hash)

	response := RPCMessage{
		Type: CliGet,
		IsNode: true,
		Sender: network.table.GetMe(),
		Payload: payload}
	response.SendResponse(conn, addr)
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

