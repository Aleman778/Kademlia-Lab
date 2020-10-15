package main

import (
	"os"
	"fmt"
	"net"
	"time"
)

type ConnectionData struct {
	rpcMsg RPCMessage
	addr *net.UDPAddr
	err error
}

type GetRPCConfig struct {
	writeCh chan<- GetRPCData
	conn *net.UDPConn
	timeout time.Duration
	verbose bool
}

type GetRPCData struct {
	rpcMsg RPCMessage
	addr *net.UDPAddr
	err error
}

type SendToStruct struct {
	writeCh chan<- *net.UDPConn
	rpcMsg RPCMessage
	address string
	verbose bool
}

type SendResponseStruct struct {
	rpcMsg RPCMessage
	conn *net.UDPConn
	address *net.UDPAddr
}

func GetRPCMessageStarter(readCh <-chan GetRPCConfig) {
	for {
		value, more := <-readCh
		if !more {
			return
		}
		go GetRPCMessage(value.writeCh, value.conn, value.timeout, value.verbose)
	}
}

func GetRPCMessage(writeCh chan<- GetRPCData, conn *net.UDPConn, timeout time.Duration, verbose bool) {
	defer close(writeCh)
	var rpcMsg RPCMessage
	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout * time.Second))
	}
	inputBytes := make([]byte, 1024)
	length, addr, err := conn.ReadFromUDP(inputBytes)
	if err != nil {
		writeCh <- GetRPCData{rpcMsg, nil, err}
		return
	}

	DecodeRPCMessage(&rpcMsg, inputBytes[:length])
	if verbose {
		fmt.Println("Recived Msg from ", addr, " :\n", rpcMsg.String())
	}

	writeCh <- GetRPCData{rpcMsg, addr, nil}
}



func SendToStarter(readCh <-chan SendToStruct) {
	for {
		value, more := <-readCh
		if !more {
			return
		}
		go SendTo(value.writeCh, value.rpcMsg, value.address, value.verbose)
	}
}

func SendTo(writeCh chan<- *net.UDPConn, rpcMsg RPCMessage, address string, verbose bool) {
	defer close(writeCh)
    
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		fmt.Println("Error: Can't resolve the udp address: ", address)
		fmt.Println("Fatal error ", err.Error())
		writeCh <- nil
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		fmt.Println("Error: Can't connect to the udp address: ", udpAddr)
		fmt.Println("Fatal error ", err.Error())
		writeCh <- nil
		os.Exit(1)
	}

	_, err = conn.Write(EncodeRPCMessage(rpcMsg))
	if err != nil {
		fmt.Println("Error: Can't write udp message")
		fmt.Println("Fatal error ", err.Error())
		writeCh <- nil
		os.Exit(1)
	}

	if verbose {
		fmt.Println("Sent Msg to ", udpAddr, " :\n", rpcMsg.String())
	}

	writeCh <- conn
}


func SendResponseStarter(readCh <-chan SendResponseStruct) {
	for {
		value, more := <-readCh
		if !more {
			return
		}
		go SendResponse(value.rpcMsg, value.conn, value.address)
	}
}

func SendResponse(rpcMsg RPCMessage, conn *net.UDPConn, address *net.UDPAddr) {
	conn.WriteToUDP(EncodeRPCMessage(rpcMsg), address)
	fmt.Println("Sent Msg to ", address, " :\n", rpcMsg.String())
}

