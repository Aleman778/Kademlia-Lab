package main

import (
	"net"
	"fmt"
)

const PORT = ":8080"

func InitServer() {
	me := NewContact(NewRandomKademliaID(), resolveHostIp())
	network := Network{NewRoutingTable(me)}
	RunServer(&network)
}

func JoinNetwork(address string) {
	me := NewContact(NewRandomKademliaID(), resolveHostIp())
	network := Network{NewRoutingTable(me)}

	network.InitNodeLookup(address)

	RunServer(&network)
}

func RunServer(network *Network) {
	network.Listen(PORT)
}

func resolveHostIp() (string) {

    netInterfaceAddresses, err := net.InterfaceAddrs()

    if err != nil { return "" }

    for _, netInterfaceAddress := range netInterfaceAddresses {

        networkIp, ok := netInterfaceAddress.(*net.IPNet)

        if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil {

            ip := networkIp.IP.String()

            fmt.Println("Resolved IP: " + ip + PORT)

	    return ip + PORT
        }
    }
    return ""
}

