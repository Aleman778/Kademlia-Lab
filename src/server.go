package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"sync"
)

const k = 5
const ALPHA = 3
const PORT = ":8080"

type Server struct {
	table *RoutingTable
	storage *Storage

	getRpcCh chan<- GetRPCConfig

	sendToCh chan<- SendToStruct
	sendResponseCh chan<- SendResponseStruct
}

func CreateServer() Server {
	me := NewContact(NewRandomKademliaID(), resolveHostIp(PORT))

	getRpcCh := make(chan GetRPCConfig)
	go GetRPCMessageStarter(getRpcCh)

	sendToCh := make(chan SendToStruct)
	go SendToStarter(sendToCh)

	sendResponseCh := make(chan SendResponseStruct)
	go SendResponseStarter(sendResponseCh)

	return Server{NewRoutingTable(me), NewStorage(), getRpcCh, sendToCh, sendResponseCh}
}

func InitServer() {
	server := CreateServer()
	go RunRestServer()
	RunServer(&server)
}

func JoinNetwork(address string) {
	server := CreateServer()

	server.BootstrapNode(address)

	RunServer(&server)
}

func RunServer(server *Server) {
	server.Listen(PORT)
}


func resolveHostIp(port string) (string) {
    netInterfaceAddresses, err := net.InterfaceAddrs()
    if err != nil { return "" }
    for _, netInterfaceAddress := range netInterfaceAddresses {
        serverIp, ok := netInterfaceAddress.(*net.IPNet)
        if ok && !serverIp.IP.IsLoopback() && serverIp.IP.To4() != nil {
            ip := serverIp.IP.String()
            fmt.Println("Resolved IP: " + ip + port)
	    return ip + port
        }
    }
    return ""
}



func (server *Server) Listen(port string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	checkError(err)

	conn, err := net.ListenUDP("udp4", udpAddr)
	checkError(err)
	defer conn.Close()

	fmt.Println("Server setup finished")

	for {
		server.HandleClient(conn)
	}
}


func (server *Server) SendPingMessage(contact *Contact) bool {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{"", nil, maxExpire, nil}}

	conn := rpcMsg.SendTo(server.sendToCh, contact.Address, true)
	defer conn.Close()

	readCh := make(chan GetRPCData)
	server.getRpcCh <- GetRPCConfig{readCh, conn, 5, true}
	data := <-readCh

	if data.err != nil {
		fmt.Println("\nGetting response timeout")
		return false
	}

	return true
}

func (server *Server) BootstrapNode(address string) {
	var called []Contact
	me := server.table.GetMe()
	me.CalcDistance(me.ID)
	called = append(called, me)

	var notCalled []Contact
	contacts, err := server.SendFindContactMessage(address, *me.ID, maxExpire)
	checkError(err)

	for _, contact := range contacts {
		if !InCandidates(notCalled, contact) && !InCandidates(called, contact) {
			contact.CalcDistance(me.ID)
			notCalled = append(notCalled, contact)
		}
	}

	sort.Sort(ByDistance(notCalled))
	server.StartNodeLookup(*me.ID, notCalled, maxExpire)
}

func (server *Server) NodeLookup(id KademliaID, expire int64) []Contact {
	var notCalled []Contact
	notCalled = append(notCalled, server.table.FindClosestContacts(&id, k)...)
	sort.Sort(ByDistance(notCalled))

	return server.StartNodeLookup(id, notCalled, expire)
}

func (server *Server) StartNodeLookup(id KademliaID, notCalled []Contact, expire int64) []Contact {
	contactsCh := make(chan LookupResponse)
	defer close(contactsCh)

	contactCh := make(chan Contact)
	defer close(contactCh)

	go server.NodeLookupSender(id, contactsCh, contactCh, expire)

	contacts, _ := RunLookup(&id, server.table.GetMe(), notCalled, contactCh, contactsCh)
	return contacts
}



func (server *Server) NodeLookupSender(id KademliaID, writeCh chan<- LookupResponse, 
    readCh <-chan Contact, expire int64) {
	for {
		contact, more := <-readCh
		if !more {
			return
		}
		go func(writeCh chan<- LookupResponse, contact Contact) {
			contacts, err := server.SendFindContactMessage(contact.Address, id, expire)
			if err != nil {
				writeCh <- LookupResponse{[]Contact{}, contact, false}
			} else {
				writeCh <- LookupResponse{contacts, contact, false}
			}
		}(writeCh, contact)
	}
}

func (server *Server) ValueLookup(hash string, expire int64) Payload {
	if val, ok := server.storage.Load(hash, expire); ok {
		return Payload{
			Hash: hash,
			Data: val,
			Contacts: nil};
	}

	id := NewHashedID(hash)
	var called []Contact
	me := server.table.GetMe()
	me.CalcDistance(&id)
	called = append(called, me)

	var notCalled []Contact
	notCalled = append(notCalled, server.table.FindClosestContacts(&id, k)...)
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

		go server.ValueLookupSender(hash, contactsCh, contactCh, intermediateCh)

		_, contact := RunLookup(&id, server.table.GetMe(), notCalled, contactCh, contactsCh)

		payload := <-inbetweenCh
        payload.TTL = expire
		go func(address string) {
			rpcMsg := RPCMessage{
				Type: Store,
				IsNode: true,
				Sender: server.table.GetMe(),
				Payload: payload}
			conn := rpcMsg.SendTo(server.sendToCh, address, true)
			defer conn.Close()
		}(contact.Address)
	}()

	payload := <-resultCh
	return payload
}

func (server *Server) ValueLookupSender(hash string, writeCh chan<- LookupResponse, readCh <-chan Contact, resultCh chan<- Payload) {
	isDone := false
	mutex := sync.RWMutex{}
	for {
		contact, more := <-readCh
		if !more {
			if !isDone {
				mutex.Lock()
				isDone = true
				mutex.Unlock()

				resultCh <- Payload{"", nil, maxExpire, nil}
			}
			return
		}
		go func(writeCh chan<- LookupResponse, contact Contact) {
			payload, err := server.SendFindDataMessage(contact.Address, hash)
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

func (server *Server) SendFindDataMessage(address string, hash string) (Payload, error) {
	rpcMsg := RPCMessage{
		Type: FindValue,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload {
			Hash: hash,
			Data: nil,
			Contacts: nil,
		}}
	conn := rpcMsg.SendTo(server.sendToCh, address, true)
	defer conn.Close()

	readCh := make(chan GetRPCData)
	server.getRpcCh <- GetRPCConfig{readCh, conn, 15, true}
	data := <-readCh
	if data.err != nil {
		fmt.Println("\nGetting response timeout")
		return data.rpcMsg.Payload, data.err
	}
	return data.rpcMsg.Payload, nil
}


func (server *Server) SendFindContactMessage(addr string, id KademliaID, expire int64) ([]Contact, error) {
	contacts := make([]Contact, 1)
	contacts[0] = NewContact(&id, "")
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{
			Hash: "",
			Data: nil,
            TTL: expire,
			Contacts: contacts,
		}}

	conn := rpcMsg.SendTo(server.sendToCh, addr, true)
	defer conn.Close()

	readCh := make(chan GetRPCData)
	server.getRpcCh <- GetRPCConfig{readCh, conn, 15, true}
	data := <-readCh

	if data.err != nil {
		fmt.Println("\nGetting response timeout")
		return []Contact{}, data.err
	}

	if data.rpcMsg.IsNode {
		server.AddContact(data.rpcMsg.Sender)
	}

	return data.rpcMsg.Payload.Contacts, nil
}

func (server *Server) HandleClient(conn *net.UDPConn) {
	readCh := make(chan GetRPCData)
	server.getRpcCh <- GetRPCConfig{readCh, conn, 0, true}
	data := <-readCh

	if data.err != nil {
		return
	}
    fmt.Printf("TTL = %d\n", data.rpcMsg.Payload.TTL)
	if data.rpcMsg.IsNode {
		server.AddContact(data.rpcMsg.Sender)
	}

	switch data.rpcMsg.Type {
	case Ping:
		go server.HandlePingMessage(conn, data.addr)
	case Store:
		go server.HandleStoreMessage(&data.rpcMsg, conn, data.addr)
	case FindNode:
		go server.HandleFindNodeMessage(&data.rpcMsg, conn, data.addr)
	case FindValue:
		go server.HandleFindValueMessage(&data.rpcMsg, conn, data.addr)
	case CliPut:
		go server.HandleCliPutMessage(&data.rpcMsg, conn, data.addr)
	case CliGet:
		go server.HandleCliGetMessage(&data.rpcMsg, conn, data.addr)
	case CliExit:
		server.HandleCliExitMessage(conn, data.addr)
	}
}

func (server *Server) HandlePingMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{"",nil, maxExpire, nil,}}

	server.sendResponseCh <- SendResponseStruct{rpcMsg, conn, addr}
}

func (server *Server) HandleStoreMessage(msg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Store,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{"", nil, maxExpire, nil}}
	server.storage.Store(msg.Payload.Hash, msg.Payload.Data, msg.Payload.TTL)
	server.sendResponseCh <- SendResponseStruct{rpcMsg, conn, addr}
}

func (server *Server) HandleFindNodeMessage(msg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	var id KademliaID = *msg.Payload.Contacts[0].ID
	contacts := server.table.FindClosestContacts(&id, k)
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{
			Hash: "",
			Data: nil,
			Contacts: contacts,
		}}
	server.sendResponseCh <- SendResponseStruct{rpcMsg, conn, addr}
}

func (server *Server) HandleFindValueMessage(msg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	if val, ok := server.storage.Load(msg.Payload.Hash, msg.Payload.TTL); ok {
		rpcMsg := RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{
				Hash: msg.Payload.Hash,
				Data: val,
				Contacts: nil,
			}}
		server.sendResponseCh <- SendResponseStruct{rpcMsg, conn, addr}
	} else {
		hashID := NewKademliaID(msg.Payload.Hash)
		closest := server.table.FindClosestContacts(hashID, k)
		rpcMsg := RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{
				Hash: "",
				Data: nil,
				Contacts: closest,
			}}
		server.sendResponseCh <- SendResponseStruct{rpcMsg, conn, addr}
	}
}

func (server *Server) HandleCliPutMessage(rpcMsg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	id := NewHashedID(rpcMsg.Payload.Hash)
	closest := server.NodeLookup(id, rpcMsg.Payload.TTL)
	for _, c := range closest {
		go func(address string) {
			rpcMsg := RPCMessage{
				Type: Store,
				IsNode: true,
				Sender: server.table.GetMe(),
				Payload: Payload {
					Hash: string(rpcMsg.Payload.Hash),
                    Data: rpcMsg.Payload.Data,
                    TTL: rpcMsg.Payload.TTL,
					Contacts: nil,
				}}
			conn := rpcMsg.SendTo(server.sendToCh, address, true)
			defer conn.Close()

			readCh := make(chan GetRPCData)
			server.getRpcCh <- GetRPCConfig{readCh, conn, 5, true}
			data := <-readCh

			if data.err != nil {
				fmt.Println("\nGeting response timeout")
			}
		}(c.Address)
	}
	response := RPCMessage{
		Type: CliPut,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{rpcMsg.Payload.Hash, nil, maxExpire, nil}}
	server.sendResponseCh <- SendResponseStruct{response, conn, addr}
}

func (server *Server) HandleCliGetMessage(rpcMsg *RPCMessage, conn *net.UDPConn, addr *net.UDPAddr) {
	payload := server.ValueLookup(rpcMsg.Payload.Hash, rpcMsg.Payload.TTL)

	response := RPCMessage{
		Type: CliGet,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: payload}
	server.sendResponseCh <- SendResponseStruct{response, conn, addr}
}

func (server *Server) HandleCliExitMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: CliExit,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{"", nil, maxExpire, nil}}
	server.sendResponseCh <- SendResponseStruct{rpcMsg, conn, addr}

	fmt.Println("Shutting down server")
	close(server.getRpcCh)
	close(server.sendToCh)
	close(server.sendResponseCh)
	os.Exit(0);
}

func (server *Server) AddContact(contact Contact) {
	bucketIndex := server.table.getBucketIndex(contact.ID)

	if server.table.IsBucketFull(bucketIndex) {
		lastContact := server.table.GetLastInBucket(bucketIndex)
		isAlive := server.SendPingMessage(&lastContact)
		if isAlive {
			return
		}

		server.table.RemoveContactFromBucket(bucketIndex, lastContact)
		server.table.AddContact(contact)
	} else {
		server.table.AddContact(contact)
	}
}

