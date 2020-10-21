package main

import "testing"

func TestRPCTypeString(t *testing.T) {
	if Ping.String() != "Ping" {
		t.Errorf("String of Ping does not return Ping")
	}
	if Store.String() != "Store" {
		t.Errorf("String of Store does not return Store")
	}
	if FindNode.String() != "FindNode" {
		t.Errorf("String of FindNode does not return FindNode")
	}
	if FindValue.String() != "FindValue" {
		t.Errorf("String of FindValue does not return FindValue")
	}
	if CliPut.String() != "CliPut" {
		t.Errorf("String of CliPut does not return CliPut")
	}
	if CliGet.String() != "CliGet" {
		t.Errorf("String of CliGet does not return CliGet")
	}
	if CliExit.String() != "CliExit" {
		t.Errorf("String of CliExit does not return CliExit")
	}
	var rpc RPCType = 8
	if rpc.String() != "" {
		t.Errorf("Empty string not returned on invalid RPCType")
	}
}

func TestRPCMessageString(t *testing.T) {
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	if rpc.String() != "\tType: Ping\n\tID: contact(\"<nil>\", \"\")" {
		t.Errorf("RPCMessage string not formatted correctly")
	}
}

func TestEncodeDecode(t *testing.T) {
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	data := EncodeRPCMessage(rpc)
	var toDecode RPCMessage
	DecodeRPCMessage(&toDecode, data)
	if rpc.Type != toDecode.Type && rpc.IsNode != toDecode.IsNode {
		t.Errorf("RPC Message encoded and decoded contains wrong values")
	}
}

func TestSendToRPC(t *testing.T) {
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	writeChan := make(chan SendToStruct, 5)
	go rpc.SendTo(writeChan, "localhost:8001", true)
}
