package main

import (
	"bytes"
	"encoding/gob"
)

type RPCMessage struct {
	Type RPCType
	Me Contact
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
	return "Type: " + msg.Type.String() + "\nID: " + msg.Me.String() + "\nData: " + string(msg.Data)
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

