package main

import (
	"fmt"
	"net"
    "os"
)

type Network struct {
}

func Listen(routingTable *RoutingTable, port string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	checkError(err)

	conn, err := net.ListenUDP("udp4", udpAddr)
	checkError(err)
	defer conn.Close()

	for {
		handleClient(routingTable, conn)
	}
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}

func NodeLookup(routingTable *RoutingTable, addr string, id KademliaID) []Contact {
	rpcMsg := RPCMessage{
		Type: FindNode,
		Me: routingTable.me,
		Data: EncodeKademliaID(id)}

	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	checkError(err)

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	checkError(err)

	defer conn.Close()

	_, err = conn.Write(EncodeRPCMessage(rpcMsg))
	checkError(err)

	inputBytes := make([]byte, 1024)
	length, _, _ := conn.ReadFromUDP(inputBytes)

	var rrpcMsg RPCMessage
	DecodeRPCMessage(&rrpcMsg, inputBytes[:length])

	routingTable.AddContact(rrpcMsg.Me)
	var contacts []Contact
	DecodeContacts(&contacts, rrpcMsg.Data)
	return contacts
}

func handleClient(routingTable *RoutingTable, conn *net.UDPConn) {

	inputBytes := make([]byte, 1024)
	length, addr, _ := conn.ReadFromUDP(inputBytes)

	var rpcMsg RPCMessage
	DecodeRPCMessage(&rpcMsg, inputBytes[:length])
	fmt.Println(rpcMsg.String())

	routingTable.AddContact(rpcMsg.Me)

	switch rpcMsg.Type {
	case Ping:
		HandlePingMessage(routingTable, rpcMsg.Data, conn, addr)
	case Store:
		HandleStoreMessage(routingTable, rpcMsg.Data, conn, addr)
	case FindNode:
		HandleFindNodeMessage(routingTable, rpcMsg.Data, conn, addr)
	case FindValue:
		HandleFindValueMessage(routingTable, rpcMsg.Data, conn, addr)
    case ExitNode:
        HandleExitNodeMessage(routingTable, conn, addr)
	}

}

func HandlePingMessage(routingTable *RoutingTable, Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Ping,
		Me: routingTable.me,
		Data: Data}
	conn.WriteToUDP(EncodeRPCMessage(rpcMsg), addr)
	//TODO
}

func HandleStoreMessage(routingTable *RoutingTable, Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: Store,
		Me: routingTable.me,
		Data: Data}
	conn.WriteToUDP(EncodeRPCMessage(rpcMsg), addr)
	//TODO
}

func HandleFindNodeMessage(routingTable *RoutingTable, Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	var id KademliaID
	DecodeKademliaID(&id, Data)
	contacts := routingTable.FindClosestContacts(&id, 3)
	rpcMsg := RPCMessage{
		Type: FindNode,
		Me: routingTable.me,
		Data: EncodeContacts(contacts)}
	conn.WriteToUDP(EncodeRPCMessage(rpcMsg), addr)
}

func HandleFindValueMessage(routingTable *RoutingTable, Data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: FindValue,
		Me: routingTable.me,
		Data: Data}
	conn.WriteToUDP(EncodeRPCMessage(rpcMsg), addr)
	//TODO
}

func HandleExitNodeMessage(routingTable *RoutingTable, conn *net.UDPConn, addr *net.UDPAddr) {
	rpcMsg := RPCMessage{
		Type: ExitNode,
		Me: routingTable.me,
		Data: nil}
	conn.WriteToUDP(EncodeRPCMessage(rpcMsg), addr)
    os.Exit(0);
}
