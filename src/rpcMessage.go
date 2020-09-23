package main

import (
	"bytes"
	"encoding/gob"
	"net"
	"os"
	"fmt"
)

type RPCMessage struct {
	Type RPCType
	IsNode bool
	Sender Contact
	Payload Payload
}

type Payload struct {
	Hash string
	Data []byte
	Contacts []Contact
}

type RPCType int

const (
	Ping RPCType = iota
	Store
	FindNode
	FindValue
    CliPut
    CliGet
    CliExit
	Test
)

func (t RPCType) String() string {
	rpcType := [...]string{"Ping", "Store", "FindNode", "FindValue", "CliPut", "CliGet", "CliExit", "Test"}
	if len(rpcType) < int(t) {
		return ""
	}

	return rpcType[t]
}


func (msg RPCMessage) String() string {
	return "\tType: " + msg.Type.String() + "\n\tID: " + msg.Sender.String()
}

func (rpcMsg RPCMessage) SendTo(address string) *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		fmt.Println("Error: Can't resolve the udp address: ", address)
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		fmt.Println("Error: Can't connect to the udp address: ", udpAddr)
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}

	_, err = conn.Write(EncodeRPCMessage(rpcMsg))
	if err != nil {
		fmt.Println("Error: Can't write udp message")
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}

	fmt.Println("Sent Msg to ", udpAddr, " :\n", rpcMsg.String())

	return conn
}

func (rpcMsg RPCMessage) SendResponse(conn *net.UDPConn, address *net.UDPAddr) {
	conn.WriteToUDP(EncodeRPCMessage(rpcMsg), address)
	fmt.Println("Sent Msg to ", address, " :\n", rpcMsg.String())
}

func EncodeRPCMessage(rpcMessage RPCMessage) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(rpcMessage)
	checkError(err)
	return buffer.Bytes()
}

func DecodeRPCMessage(rpcMessage *RPCMessage, data []byte) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(rpcMessage)
	checkError(err)
}

