package main

import (
	"fmt"
	"net"
	"os"
	"encoding/gob"
	"bytes"
)

type RPCMessage struct {
	Type RPCType
	Data []byte
}

type RPCType int

const (
	Ping RPCType = iota
	Store
	FindNode
	FindValue
)

func (t RPCType) String() string {
	rpcType := [...]string{"Ping", "Store", "FindNode", "FindValue"}
	if len(rpcType) < int(t) {
		return ""
	}

	return rpcType[t]
}


func (msg RPCMessage) String() string {
	return "Type: " + msg.Type.String() + "\nData: " + string(msg.Data)
}


type RPCFindNode [20]byte


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
	rpcType := RPCMessage{
		Type: rpc,
		Data: []byte("asdasda")}

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	checkError(err)

	defer conn.Close()

	for i := 0; i < 3; i++ {
		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		decoder := gob.NewDecoder(&buffer)

		err = encoder.Encode(rpcType)
		checkError(err)

		_, err = conn.Write(buffer.Bytes())
		checkError(err)

		inputBytes := make([]byte, 1024)
		length, _, err := conn.ReadFromUDP(inputBytes)
		buffer.Write(inputBytes[:length])

		var rrpcType RPCMessage
		err = decoder.Decode(&rrpcType)
		checkError(err)

		fmt.Println(rrpcType.String())
		buffer.Reset()
	}
}



func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

