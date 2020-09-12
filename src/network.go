package main

import (
	"fmt"
	"net"
	"bytes"
	"encoding/gob"
)

type Network struct {
}

func Listen(port string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	checkError(err)

	conn, err := net.ListenUDP("udp4", udpAddr)
	checkError(err)
	defer conn.Close()

	for {
		handleClient(conn)
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

func handleClient(conn *net.UDPConn) {

	inputBytes := make([]byte, 1024)
	length, addr, err := conn.ReadFromUDP(inputBytes)
	buffer := bytes.NewBuffer(inputBytes[:length])

	decoder := gob.NewDecoder(buffer)

	var rpcType RPCMessage
	err = decoder.Decode(&rpcType)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		return

	}
	fmt.Println(rpcType.String())

	switch rpcType.Type {
	case Ping:
		HandlePingMessage(rpcType.Data, addr)
	case Store:
		HandleStoreMessage(rpcType.Data, addr)
	case FindNode:
		HandleFindNodeMessage(rpcType.Data, addr)
	case FindValue:
		HandleFindValueMessage(rpcType.Data, addr)
	}

	encoder := gob.NewEncoder(buffer)
	err = encoder.Encode(rpcType)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		return

	}

	conn.WriteToUDP(buffer.Bytes(), addr)
}

func HandlePingMessage(Data []byte, addr *net.UDPAddr) {
	//TODO
}

func HandleStoreMessage(Data []byte, addr *net.UDPAddr) {
	//TODO
}

func HandleFindNodeMessage(Data []byte, addr *net.UDPAddr) {
	//TODO
}

func HandleFindValueMessage(Data []byte, addr *net.UDPAddr) {
	//TODO
}

