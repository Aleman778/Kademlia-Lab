package main

import (
	"fmt"
	"net"
	"os"
	"encoding/gob"
	"bytes"
)

type RPCMessage struct {
	Type string
	Data []byte
}


func (msg RPCMessage) String() string {
	return "Type: " + msg.Type + "\nData: " + string(msg.Data)
}


func main() {

	udpAddr, err := net.ResolveUDPAddr("udp4", ":8080")
	checkError(err)

	conn, err := net.ListenUDP("udp4", udpAddr)
	checkError(err)
	defer conn.Close()

	for {
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {

	inputBytes := make([]byte, 1024)
	length, addr, err := conn.ReadFromUDP(inputBytes)
	buffer := bytes.NewBuffer(inputBytes[:length])

	decoder := gob.NewDecoder(buffer)

	var person RPCMessage
	err = decoder.Decode(&person)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		return

	}
	fmt.Println(person.String())

	encoder := gob.NewEncoder(buffer)
	err = encoder.Encode(person)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		return

	}

	conn.WriteToUDP(buffer.Bytes(), addr)
}


func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
