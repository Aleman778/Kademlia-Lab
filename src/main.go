package main

import (
	"fmt"
	"net"
	"os"
)


func main() {
	if len(os.Args) != 3 {
		InitServer()
	} else {
		var rpcType RPCType

		switch os.Args[2] {
		case "ping":
			rpcType = Ping
		case "store":
			rpcType = Store
		case "find-node":
			rpcType = FindNode
		case "find-value":
			rpcType = FindValue
		}
		client(os.Args[1], rpcType)
	}
}

func client(service string, rpc RPCType) {
	rpcMsg := RPCMessage{
		Type: rpc,
		Data: []byte(nil)}

	switch rpcMsg.Type {
	case Ping:
	case Store:
	case FindNode:
		rpcMsg.Data = EncodeKademliaID(*NewRandomKademliaID())
	case FindValue:
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
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
	fmt.Println(rrpcMsg.String())
}


func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

