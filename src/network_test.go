package main

import (
	"testing"
	"net"
	"time"
	"strconv"
	"crypto/sha1"
	"encoding/hex"
)


func TestSendPingMessage(t *testing.T) {
	network := createNetwork()
	readCh := make(chan RPCMessage)
	address := ":7070"

	go setupListner(readCh, address)

	time.Sleep(1 * time.Second)

	contact := NewContact(NewRandomKademliaID(), address)
	network.SendPingMessage(&contact)

	rpcMsg, more := <-readCh
	if !more {
		t.Error("Chanal unexpected close")
	}

	verifyRPCMessage(t, rpcMsg, Ping, true, network.table.me)
	verifyPayload(t, rpcMsg.Payload, "", nil, nil)
}

func TestSendFindContactMessage(t *testing.T) {
	network := createNetwork()
	readCh := make(chan RPCMessage)
	address := ":7070"

	go setupListner(readCh, address)

	time.Sleep(1 * time.Second)

	contact := NewContact(NewRandomKademliaID(), address)
	network.SendFindContactMessage(contact.Address, *contact.ID)

	rpcMsg, more := <-readCh
	if !more {
		t.Error("Chanal unexpected close")
	}

	verifyRPCMessage(t, rpcMsg, FindNode, true, network.table.me)
	verifyPayload(t, rpcMsg.Payload, "", nil, []Contact{contact})
}

















func TestHandleClientPing(t *testing.T) {
	finished := make(chan bool)
	network := createNetwork()
	timeout := time.Duration(5)

	go func(netowrk Network) {
		udpAddr, err := net.ResolveUDPAddr("udp4", network.table.me.Address)
		checkError(err)

		conn, err := net.ListenUDP("udp4", udpAddr)
		checkError(err)
		defer conn.Close()

		network.HandleClient(conn)
		time.Sleep((timeout + 1) * time.Second)
		finished <- true
	}(*network)

	time.Sleep(1 * time.Second)

	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: network.table.me,
		Payload: Payload{"", nil, nil}}


	conn, err := sendTo(rpcMsg, network.table.me.Address)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, timeout, true)
	if err != nil {
		t.Error(err)
		return
	}

	verifyRPCMessage(t, response, Ping, true, network.table.me)
	verifyPayload(t, response.Payload, "", nil, nil)

	<-finished
}

func TestHandleClientStore(t *testing.T) {
	finished := make(chan bool)
	network := createNetwork()
	timeout := time.Duration(5)

	go func(netowrk Network) {
		udpAddr, err := net.ResolveUDPAddr("udp4", network.table.me.Address)
		checkError(err)

		conn, err := net.ListenUDP("udp4", udpAddr)
		checkError(err)
		defer conn.Close()

		network.HandleClient(conn)
		time.Sleep((timeout + 1) * time.Second)
		finished <- true
	}(*network)

	time.Sleep(1 * time.Second)

	hash := sha1.Sum([]byte("test"))
	hash_string := hex.EncodeToString(hash[:])
	rpcMsg := RPCMessage{
		Type: Store,
		IsNode: true,
		Sender: network.table.me,
		Payload: Payload{hash_string, []byte("test"), nil}}


	conn, err := sendTo(rpcMsg, network.table.me.Address)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, timeout, true)
	if err != nil {
		t.Error(err)
		return
	}

	verifyRPCMessage(t, response, Store, true, network.table.me)
	verifyPayload(t, response.Payload, "", nil, nil)

	<-finished
}



func TestHandleClientFindNode(t *testing.T) {
	finished := make(chan bool)
	network := createNetwork()
	timeout := time.Duration(5)

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")

	network.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	network.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"))
	network.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"))
	network.AddContact(contact3)
	network.AddContact(contact2)
	network.AddContact(contact1)

	go func(netowrk Network) {
		udpAddr, err := net.ResolveUDPAddr("udp4", network.table.me.Address)
		checkError(err)

		conn, err := net.ListenUDP("udp4", udpAddr)
		checkError(err)
		defer conn.Close()

		network.HandleClient(conn)
		time.Sleep((timeout + 1) * time.Second)
		finished <- true
	}(*network)

	time.Sleep(timeout * time.Second)

	hash := sha1.Sum([]byte("test"))
	hash_string := hex.EncodeToString(hash[:])
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: network.table.me,
		Payload: Payload{hash_string, nil, []Contact{contact1}}}


	conn, err := sendTo(rpcMsg, network.table.me.Address)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, 15, true)
	if err != nil {
		t.Error(err)
		return
	}

	response.Payload.Contacts = response.Payload.Contacts[:3]

	verifyRPCMessage(t, response, FindNode, true, network.table.me)
	verifyPayload(t, response.Payload, "", nil, []Contact{contact1, contact2, contact3})

	<-finished
}

func TestHandleClientFindValue(t *testing.T) {
	finished := make(chan bool)
	network := createNetwork()
	timeout := time.Duration(5)

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")

	network.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	network.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"))
	network.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"))
	network.AddContact(contact3)
	network.AddContact(contact2)
	network.AddContact(contact1)

	go func(netowrk Network) {
		udpAddr, err := net.ResolveUDPAddr("udp4", network.table.me.Address)
		checkError(err)

		conn, err := net.ListenUDP("udp4", udpAddr)
		checkError(err)
		defer conn.Close()

		network.HandleClient(conn)
		time.Sleep((timeout + 1) * time.Second)
		finished <- true
	}(*network)

	time.Sleep(timeout * time.Second)

	rpcMsg := RPCMessage{
		Type: FindValue,
		IsNode: true,
		Sender: network.table.me,
		Payload: Payload{"2111111400000000000000000000000000000000", nil, []Contact{contact1}}}


	conn, err := sendTo(rpcMsg, network.table.me.Address)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, 15, true)
	if err != nil {
		t.Error(err)
		return
	}

	response.Payload.Contacts = response.Payload.Contacts[:3]

	verifyRPCMessage(t, response, FindValue, true, network.table.me)
	verifyPayload(t, response.Payload, "", nil, []Contact{contact1, contact2, contact3})

	<-finished
}


func TestHandleClientFindValueFail(t *testing.T) {
	finished := make(chan bool)
	network := createNetwork()
	timeout := time.Duration(5)

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	network.storage.Store(hash_string, []byte(data))

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")

	network.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	network.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"))
	network.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"))
	network.AddContact(contact3)
	network.AddContact(contact2)
	network.AddContact(contact1)

	go func(netowrk Network) {
		udpAddr, err := net.ResolveUDPAddr("udp4", network.table.me.Address)
		checkError(err)

		conn, err := net.ListenUDP("udp4", udpAddr)
		checkError(err)
		defer conn.Close()

		network.HandleClient(conn)
		time.Sleep((timeout + 1) * time.Second)
		finished <- true
	}(*network)

	time.Sleep(timeout * time.Second)

	rpcMsg := RPCMessage{
		Type: FindValue,
		IsNode: true,
		Sender: network.table.me,
		Payload: Payload{hash_string, []byte(data), []Contact{contact1}}}


	conn, err := sendTo(rpcMsg, network.table.me.Address)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, 15, true)
	if err != nil {
		t.Error(err)
		return
	}

	verifyRPCMessage(t, response, FindValue, true, network.table.me)
	verifyPayload(t, response.Payload, hash_string, []byte(data), nil)

	<-finished
}

func TestHandleClientCliPut(t *testing.T) {
	finished := make(chan bool)
	network := createNetwork()
	timeout := time.Duration(10)

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	address := "localhost:8002"
	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), address)
	network.AddContact(contact1)
	go setupReplier(address)

	go func(netowrk Network) {
		udpAddr, err := net.ResolveUDPAddr("udp4", network.table.me.Address)
		checkError(err)

		conn, err := net.ListenUDP("udp4", udpAddr)
		checkError(err)
		defer conn.Close()

		network.HandleClient(conn)
		time.Sleep((timeout + 1) * time.Second)
		finished <- true
	}(*network)

	time.Sleep(timeout * time.Second)

	rpcMsg := RPCMessage{
		Type: CliPut,
		IsNode: true,
		Sender: network.table.me,
		Payload: Payload{hash_string, []byte(data), nil}}


	conn, err := sendTo(rpcMsg, network.table.me.Address)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, timeout*2, true)
	if err != nil {
		t.Error(err)
		return
	}

	verifyRPCMessage(t, response, CliPut, true, network.table.me)
	verifyPayload(t, response.Payload, hash_string, nil, nil)

	<-finished
}


func TestHandleClientCliGet(t *testing.T) {
	finished := make(chan bool)
	network := createNetwork()
	timeout := time.Duration(10)

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	address := "localhost:8002"
	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), address)
	network.AddContact(contact1)
	go setupReplier(address)

	go func(netowrk Network) {
		udpAddr, err := net.ResolveUDPAddr("udp4", network.table.me.Address)
		checkError(err)

		conn, err := net.ListenUDP("udp4", udpAddr)
		checkError(err)
		defer conn.Close()

		network.HandleClient(conn)
		time.Sleep((timeout + 1) * time.Second)
		finished <- true
	}(*network)

	time.Sleep(timeout * time.Second)

	rpcMsg := RPCMessage{
		Type: CliGet,
		IsNode: true,
		Sender: network.table.me,
		Payload: Payload{hash_string, []byte(data), nil}}


	conn, err := sendTo(rpcMsg, network.table.me.Address)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	response, _, err := GetRPCMessage(conn, timeout*2, true)
	if err != nil {
		t.Error(err)
		return
	}

	verifyRPCMessage(t, response, CliGet, true, network.table.me)
	verifyPayload(t, response.Payload, "", nil, nil)

	<-finished
}


func TestNetworkAddContact(t *testing.T) {
	network := createNetwork()
	removed := NewContact(NewKademliaID("00000000000000000000000000000000000000" + strconv.Itoa(bucketSize - 1)), "localhost:8002")
	for i := 0; i <= bucketSize; i++ {
		var contact Contact
		if i > 9 {
			contact = NewContact(NewKademliaID("00000000000000000000000000000000000000" + strconv.Itoa(i)), "localhost:8002")
		} else {
			contact = NewContact(NewKademliaID("000000000000000000000000000000000000000" + strconv.Itoa(i)), "localhost:8002")
		}

		network.AddContact(contact)
	}

	contacts := network.table.FindClosestContacts(removed.ID, 1)
	if contacts[0].ID.Equals(removed.ID) {
		t.Errorf("Expected %s to not be in RoutingTable", removed.String())
	}
}

func sendTo(rpcMsg RPCMessage, address string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(60 * time.Second)
	err = conn.SetWriteDeadline(deadline)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(EncodeRPCMessage(rpcMsg))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func setupListner(writeCh chan<- RPCMessage, address string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	checkError(err)

	conn, err := net.ListenUDP("udp4", udpAddr)
	checkError(err)
	defer conn.Close()

	rpcMsg, _, err := GetRPCMessage(conn, 0, true)

	writeCh <- rpcMsg
}

func setupReplier(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	checkError(err)

	conn, err := net.ListenUDP("udp4", udpAddr)
	checkError(err)
	defer conn.Close()

	rpcMsg, addr, err := GetRPCMessage(conn, 0, true)

	rpcMsg.SendResponse(conn, addr)
}

func createNetwork() *Network {
	me := NewContact(NewRandomKademliaID(), ":6060")
	network := Network{NewRoutingTable(me), NewStorage()}
	return &network
}

func verifyRPCMessage(t *testing.T, rpcMsg RPCMessage, expType RPCType, expIsNode bool, expSender Contact) {
	if rpcMsg.Type != expType {
		t.Errorf("Expected RPCMsg type %d got %d", expType, rpcMsg.Type)
	}

	if rpcMsg.IsNode != expIsNode {
		t.Errorf("Expected RPCMsg IsNode %t got %t", expIsNode, rpcMsg.IsNode)
	}

	if !rpcMsg.Sender.ID.Equals(expSender.ID) {
		t.Error("Expected RPCMsg Sender \n", expSender.String(), "\n got\n", rpcMsg.Sender.String())
	}
}

func verifyPayload(t *testing.T, payload Payload, expHash string, expData []byte, expContacts []Contact) {
	if payload.Hash != expHash {
		t.Errorf("Expected hash %s got %s", expHash, payload.Hash)
	}

	if len(payload.Data) != len(expData) {
		t.Errorf("Expected data %s got %s", expData, payload.Data)
	}

	for i := 0; i < len(expData); i++ {
		if payload.Data[i] != expData[i] {
			t.Errorf("Expected data %s got %s", expData, payload.Data)
		}
	}

	if len(payload.Contacts) != len(expContacts) {
		t.Error("Expected ",  len(expContacts), " contacts got ", len(payload.Contacts))
	}

	for i := 0; i < len(expContacts); i++ {
		if !payload.Contacts[i].ID.Equals(expContacts[i].ID) {
			t.Error("Expected contacts ", expContacts, " got ", payload.Contacts)
		}
	}
}

