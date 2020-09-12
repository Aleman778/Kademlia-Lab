/* Gob EchoClient
 */
package main

import (
	"fmt"
	"net"
	"os"
	"encoding/gob"
	"log"
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
	person := RPCMessage{
		Type: "Test",
		Data: []byte("asdasda")}

	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "host:port")
		os.Exit(1)
	}
	service := os.Args[1]

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	checkError(err)

	defer conn.Close()


	for i := 0; i < 3; i++ {
		var buffer bytes.Buffer
		encoder := gob.NewEncoder(&buffer)
		decoder := gob.NewDecoder(&buffer)

		err = encoder.Encode(person)
		checkError(err)

		_, err = conn.Write(buffer.Bytes())
		checkError(err)

		inputBytes := make([]byte, 1024)
		length, _, err := conn.ReadFromUDP(inputBytes)
		buffer.Write(inputBytes[:length])


		var rperson RPCMessage
		err = decoder.Decode(&rperson)
		checkError(err)

		fmt.Println(rperson.String())
		buffer.Reset()
	}

	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}


// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

