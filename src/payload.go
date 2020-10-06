package main

import (
	"encoding/gob"
	"fmt"
	"net"
)

type Payload struct {
	Msg      RPCType
	Hash     string
	Data     []byte
	Contacts []Contact
	Me       *Contact
}

type RPCType int

const (
	PING RPCType = iota
	STORE
	FINDNODE
	FINDDATA
)

func sendPayload(address string, payload *Payload) {
	ContactAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Errorf("Could not resolve contact address: ", address)
		return
	}
	/*LocalAddr, err := net.ResolveUDPAddr("udp", network.table.Me.Address)
	if err != nil {
		fmt.Errorf("Could not resolve local address: ", network.table.Me.Address)
		return
	}*/
	conn, err := net.DialUDP("udp", nil, ContactAddress)
	if err != nil {
		fmt.Errorf("Could not dial to: ", address)
		return
	}
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(*payload)
}

func sendAndRecvPayload(address string, payload *Payload) (*Payload, error) {
	ContactAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Errorf("Could not resolve contact address: ", address)
		return nil, nil
	}
	/*LocalAddr, err := net.ResolveUDPAddr("udp", network.table.Me.Address)
	if err != nil {
		fmt.Errorf("Could not resolve local address: ", network.table.Me.Address)
		return
	}*/
	conn, err := net.DialUDP("udp", nil, ContactAddress)
	if err != nil {
		fmt.Errorf("Could not dial to: ", address)
		return nil, nil
	}
	defer conn.Close()
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(payload)

	var recv Payload
	decoder := gob.NewDecoder(conn)
	err = decoder.Decode(recv)

	return &recv, nil
}
