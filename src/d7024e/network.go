package d7024e

import (
	"net"
	"crypto/sha1"
	"fmt"
	"encoding/gob"
)

const k = 5

type Network struct {
	table *RoutingTable
}

type Rpc int

const (
	findData Rpc = iota
	sendData
)

type MSG struct {
	Msg Rpc
	Data []byte
	Me Contact
}

func (network *Network) Listen(ip string, port int) {
	ln, err := net.Listen("udp", ip + string(port))

	if err != nil {
		fmt.Errorf("Error in net.Listen: %v", err)
	} else {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Errorf("Error in ln.Accept: %v", err)
		} else {
			var msg MSG
			dec := gob.NewDecoder(conn)
			err := dec.Decode(&msg)
			if err != nil {
				fmt.Errorf("Error in dec.Decode: %v", err)
			} else {
				switch msg.Msg {
				case findData:
					//TODO
				case sendData:
					//TODO
				}
			}
		}
	}
}

func (network *Network) SendPingMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

// Contact k closest nodes (using hash in distance measurement) to find data
func (network *Network) SendFindDataMessage(hash string) {
	hashID := NewKademliaID(hash)
	closest := network.table.FindClosestContacts(hashID, k)

	for _, contact := range closest {
		go func(address string) {
			conn, err := net.Dial("udp", address)
			defer conn.Close()
			if err != nil {
				fmt.Errorf("Error in SendFindDataMessage: %v", err)
			} else {
				msg := MSG{findData, [], network.table.me}
				enc := gob.NewEncoder(conn)
				err := enc.Encode(msg)
			}
		}(contact.Address)
	}
}

// Find the closest k nodes and store the data in those nodes
func (network *Network) SendStoreMessage(data []byte) {
	hash := sha1.Sum(data)
	hashID := NewKademliaID(string(hash[:]))
	closest := network.table.FindClosestContacts(hashID, k)

	for _, contact := range closest {
		go func(address string) {
			conn, err := net.Dial("udp", address) //This might go outside go func
			defer conn.Close()
			if err != nil {
				fmt.Errorf("Error in net.Dial: %v", err)
			} else {
				msg := MSG{sendData, data, network.table.me}
				enc := gob.NewEncoder(conn)
				err := enc.Encode(msg)
				if err != nil {
					fmt.Errorf("Error in enc.Encode: %v", err)
				}
			}
		}(contact.Address)
	}
}
