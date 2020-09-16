package main

import (
	"net"
	"log"
)

func InitServer() {
	me := NewContact(NewRandomKademliaID(), GetOutboundIP())
	network := Network{NewRoutingTable(me)}
	RunServer(&network)
}

func JoinNetwork(addr string) {
	me := NewContact(NewRandomKademliaID(), GetOutboundIP())
	network := Network{NewRoutingTable(me)}

	contacts := network.NodeLookup(addr, *network.table.me.ID)
	for _, contact := range contacts {
		network.table.AddContact(contact)
	}
	RunServer(&network)
}

func RunServer(network *Network) {
	network.Listen(":8080")
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

