package main

import (
	"bytes"
	"encoding/gob"
	"net"
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
    TTL int64
	Contacts []Contact
}

type RPCType int

const (
    Ping RPCType = iota
    Store
    Refresh
    FindNode
    FindValue
    CliPut
    CliGet
    CliForget
    CliExit
)

func (t RPCType) String() string {
    rpcType := [...]string{"Ping", "Store", "Refresh", "FindNode", "FindValue",
                           "CliPut", "CliGet", "CliForget", "CliExit"}
	if len(rpcType) < int(t) {
		return ""
	}

	return rpcType[t]
}


func (msg RPCMessage) String() string {
	return "\tType: " + msg.Type.String() + "\n\tID: " + msg.Sender.String()
}

func (msg RPCMessage) SendTo(writeCh chan<- SendToStruct, address string, verbose bool) *net.UDPConn {
	readCh := make(chan *net.UDPConn)
	writeCh <- SendToStruct{readCh, msg, address, verbose}
	conn := <-readCh
	return conn
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

