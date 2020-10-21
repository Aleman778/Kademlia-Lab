package main

import (
	"net"
	"testing"
	"time"
)

func TestGetRPCMessageStarter(t *testing.T) {
	rpcChan := make(chan GetRPCConfig, 5)
	writeChan := make(chan GetRPCData, 5)
	addr, err := net.ResolveUDPAddr("udp", "localhost:8001")
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
	}
	rpcConfig := GetRPCConfig{writeChan, conn, time.Until(time.Now().Add(time.Second * 10)), true}
	rpcChan <- rpcConfig
	go GetRPCMessageStarter(rpcChan)
	res := <-writeChan
	msg := res.rpcMsg
	if msg.Type != Ping && msg.IsNode != false {
		t.Errorf("Wrong RPC Message received")
	}
	close(rpcChan)
	go GetRPCMessageStarter(rpcChan)
}

func TestGetRPCMessage(t *testing.T) {
	writeChan := make(chan GetRPCData, 5)
	saddr, err := net.ResolveUDPAddr("udp", "localhost:8001")
	if err != nil {
		t.Errorf("Test failed due to error in resolving address for server")
		return
	}
	caddr, err := net.ResolveUDPAddr("udp", "localhost:8002")
	if err != nil {
		t.Errorf("Test failed due to error in resolving address for client")
		return
	}
	server, err := net.DialUDP("udp", caddr, saddr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
		return
	}
	client, err := net.DialUDP("udp", saddr, caddr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection to server")
		return
	}
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	_, err = client.Write(EncodeRPCMessage(rpc))
	if err != nil {
		t.Errorf("Could not send value")
	}
	go GetRPCMessage(writeChan, server, time.Until(time.Now().Add(time.Second*1)), true)
	res := <-writeChan
	msg := res.rpcMsg
	if msg.Type != Ping && msg.IsNode != false {
		t.Errorf("Wrong RPC Message received")
	}
}

func TestSendToStarter(t *testing.T) {
	saddr, err := net.ResolveUDPAddr("udp", "localhost:19009")
	if err != nil {
		t.Errorf("Test failed due to error in resolving address for server")
		return
	}
	server, err := net.ListenUDP("udp", saddr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
		return
	}
	defer server.Close()

	readChan := make(chan SendToStruct, 5)
	writeChan := make(chan *net.UDPConn, 5)
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	sendTo := SendToStruct{writeChan, rpc, "localhost:19009", true}
	readChan <- sendTo
	go SendToStarter(readChan)

	buffer := make([]byte, 1024)
	num, _, err := server.ReadFromUDP(buffer)
	if num == 0 || err != nil {
		t.Errorf("No data received from send response")
	}
	var msg RPCMessage
	DecodeRPCMessage(&msg, buffer)
	if msg.Type != Ping && msg.IsNode != false {
		t.Errorf("Msg sent is not correct")
	}

	close(readChan)
	go SendToStarter(readChan)
}

func TestSendTo(t *testing.T) {
	saddr, err := net.ResolveUDPAddr("udp", "localhost:19008")
	if err != nil {
		t.Errorf("Test failed due to error in resolving address for server")
		return
	}
	server, err := net.ListenUDP("udp", saddr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
		return
	}
	defer server.Close()

	writeChan := make(chan *net.UDPConn, 5)
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	address := "localhost:19008"
	go SendTo(writeChan, rpc, address, true)

	buffer := make([]byte, 1024)
	num, _, err := server.ReadFromUDP(buffer)
	if num == 0 || err != nil {
		t.Errorf("No data received from send response")
	}
	var msg RPCMessage
	DecodeRPCMessage(&msg, buffer)
	if msg.Type != Ping && msg.IsNode != false {
		t.Errorf("Msg sent is not correct")
	}

	writeChan = make(chan *net.UDPConn, 5)
	address = "idjsafhjldsfh"
	go SendTo(writeChan, rpc, address, true)
}

func TestSendResponseStarter(t *testing.T) {
	saddr, err := net.ResolveUDPAddr("udp", "localhost:19010")
	if err != nil {
		t.Errorf("Test failed due to error in resolving address for server")
		return
	}
	server, err := net.ListenUDP("udp", saddr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
		return
	}
	defer server.Close()

	readChan := make(chan SendResponseStruct, 5)
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	addr, err := net.ResolveUDPAddr("udp", "localhost:19010")
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
	}
	sendResponse := SendResponseStruct{rpc, conn, addr}
	readChan <- sendResponse
	go SendResponseStarter(readChan)

	buffer := make([]byte, 1024)
	num, _, err := server.ReadFromUDP(buffer)
	if num == 0 || err != nil {
		t.Errorf("No data received from send response")
	}
	var msg RPCMessage
	DecodeRPCMessage(&msg, buffer)
	if msg.Type != Ping && msg.IsNode != false {
		t.Errorf("Msg sent is not correct")
	}

	close(readChan)
	go SendResponseStarter(readChan)
}

func TestSendResponse(t *testing.T) {
	saddr, err := net.ResolveUDPAddr("udp", "localhost:19003")
	if err != nil {
		t.Errorf("Test failed due to error in resolving address for server")
		return
	}
	caddr, err := net.ResolveUDPAddr("udp", "localhost:19004")
	if err != nil {
		t.Errorf("Test failed due to error in resolving address for client")
		return
	}
	server, err := net.ListenUDP("udp", saddr)
	if err != nil {
		t.Errorf("Test failed due to error in creating connection")
		return
	}
	defer server.Close()
	client, err := net.DialUDP("udp", caddr, saddr)
	defer client.Close()

	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	SendResponse(rpc, client, saddr)

	buffer := make([]byte, 1024)
	num, _, err := server.ReadFromUDP(buffer)
	if num == 0 || err != nil {
		t.Errorf("No data received from send response")
	}
	var msg RPCMessage
	DecodeRPCMessage(&msg, buffer)
	if msg.Type != Ping && msg.IsNode != false {
		t.Errorf("Msg sent is not correct")
	}
}
