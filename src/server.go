package main

import (
	"net"
	"log"
)

func InitServer() {
	me := NewContact(NewRandomKademliaID(), GetOutboundIP())
	routingTable := NewRoutingTable(me)
	RunServer(routingTable)
}

func RunServer(routingTable *RoutingTable) {
	Listen(routingTable, ":8080")
}


// Get preferred outbound ip of this machine
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

