package main

import (
	"fmt"
	"net"
	"os"
	"crypto/sha1"
	"encoding/gob"
)

const k = 5

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

	for {
		network.HandleClient(conn)
	}
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	hashID := NewKademliaID(hash)
	closest := network.table.FindClosestContacts(hashID, k)

	for _, contact := range closest {
		go func(address string) {
			conn, err := net.Dial("udp", address)
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
			conn, err := net.Dial("udp", address) //This might go outside go func
			defer conn.Close()
			if err != nil {
				fmt.Errorf("Error in net.Dial: %v", err)
			} else {
				msg := MSG{sendData, data, network.table.me}
				enc := gob.NewEncoder(conn)
				err := enc.Encode(msg)
				if err != nil {
					fmt.Errorf("Error in enc.Encode: %v", err)
				}
			}
		}(contact.Address)
	}
}



func (network *Network) NodeLookup(addr string, id KademliaID) []Contact {
	rpcMsg := RPCMessage{
		Type: FindNode,
		IsNode: true,
		Sender: network.table.me,
		Data: EncodeKademliaID(id)}

	conn := rpcMsg.SendTo(addr)
	defer conn.Close()

	responesMsg, _ := GetRPCMessage(conn)

	if responesMsg.IsNode {
		network.table.AddContact(responesMsg.Sender)
	}

	var contacts []Contact
	DecodeContacts(&contacts, responesMsg.Data)
	return contacts
}

func (network *Network) HandleClient(conn *net.UDPConn) {
	rpcMsg, addr := GetRPCMessage(conn)

	network.table.AddContact(rpcMsg.Sender)

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
	}

}

func (network *Network) HandlePingMessage(Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Ping,
		IsNode: true,
		Sender: network.table.me,
		Data: Data}
	rpcMsg.SendResponse(conn, addr)
}

func (network *Network) HandleStoreMessage(Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Store,
		IsNode: true,
		Sender: network.table.me,
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
		Sender: network.table.me,
		Data: EncodeContacts(contacts)}
	rpcMsg.SendResponse(conn, addr)
}

func (network *Network) HandleFindValueMessage(Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: FindValue,
		IsNode: true,
		Sender: network.table.me,
		Data: Data}
	rpcMsg.SendResponse(conn, addr)
	//TODO
}

func (network *Network) HandleExitNodeMessage(conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: ExitNode,
		IsNode: true,
		Sender: network.table.me,
		Data: nil}
	rpcMsg.SendResponse(conn, addr)

	fmt.Println("Shutting down server")
	os.Exit(0);
}

