package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestGetRPCMessageStarter(t *testing.T) {
	rpcChan := make(chan GetRPCConfig, 5)
	writeChan := make(chan GetRPCData, 5)
	addr, err := net.ResolveUDPAddr("udp", "localhost:8001")
	if err != nil {
		fmt.Println("Test failed due to error in creating connection")
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Test failed due to error in creating connection")
	}
	rpcConfig := GetRPCConfig{writeChan, conn, time.Until(time.Now().Add(time.Second * 10)), true}
	rpcChan <- rpcConfig
	go GetRPCMessageStarter(rpcChan)
	close(rpcChan)
	go GetRPCMessageStarter(rpcChan)
}

func TestGetRPCMessage(t *testing.T) {
	writeChan := make(chan GetRPCData, 5)
	saddr, err := net.ResolveUDPAddr("udp", "localhost:8001")
	if err != nil {
		fmt.Println("Test failed due to error in resolving address for server")
		return
	}
	caddr, err := net.ResolveUDPAddr("udp", "localhost:8002")
	if err != nil {
		fmt.Println("Test failed due to error in resolving address for client")
		return
	}
	server, err := net.DialUDP("udp", caddr, saddr)
	if err != nil {
		fmt.Println("Test failed due to error in creating connection")
		return
	}
	client, err := net.DialUDP("udp", saddr, caddr)
	if err != nil {
		fmt.Println("Test failed due to error in creating connection to server")
		return
	}
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	go client.Write(EncodeRPCMessage(rpc))
	go GetRPCMessage(writeChan, server, time.Until(time.Now().Add(time.Second*1)), true)
}

func TestSendToStarter(t *testing.T) {
	readChan := make(chan SendToStruct, 5)
	writeChan := make(chan *net.UDPConn, 5)
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	sendTo := SendToStruct{writeChan, rpc, "localhost:8001", true}
	readChan <- sendTo
	go SendToStarter(readChan)
	close(readChan)
	go SendToStarter(readChan)
}

func TestSendTo(t *testing.T) {
	writeChan := make(chan *net.UDPConn, 5)
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	address := "localhost:8001"
	go SendTo(writeChan, rpc, address, true)
	writeChan = make(chan *net.UDPConn, 5)
	address = "idjsafhjldsfh"
	go SendTo(writeChan, rpc, address, true)
}

func TestSendResponseStarter(t *testing.T) {
	readChan := make(chan SendResponseStruct, 5)
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	addr, err := net.ResolveUDPAddr("udp", "localhost:8001")
	if err != nil {
		fmt.Println("Test failed due to error in creating connection")
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Test failed due to error in creating connection")
	}
	sendResponse := SendResponseStruct{rpc, conn, addr}
	readChan <- sendResponse
	go SendResponseStarter(readChan)
	close(readChan)
	go SendResponseStarter(readChan)
}

func TestSendResponse(t *testing.T) {
	saddr, err := net.ResolveUDPAddr("udp", "localhost:8003")
	if err != nil {
		fmt.Println("Test failed due to error in resolving address for server")
	}
	caddr, err := net.ResolveUDPAddr("udp", "localhost:8004")
	if err != nil {
		fmt.Println("Test failed due to error in resolving address for client")
	}
	/*server, err := net.DialUDP("udp", caddr, saddr)
	if err != nil {
		fmt.Println("Test failed due to error in creating connection")
	}
	*/
	client, err := net.DialUDP("udp", saddr, caddr)
	if err != nil {
		fmt.Println("Test failed due to error in creating connection to server")
	}
	rpc := RPCMessage{Ping, false, Contact{}, Payload{}}
	SendResponse(rpc, client, saddr)
}
