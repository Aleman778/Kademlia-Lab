package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"testing"
)

func TestNetwork(t *testing.T) {
	send := make(chan *Payload)
	defer close(send)
	contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8080")
	toContact := NewContact(NewRandomKademliaID(), "127.0.0.2:8080")
	fmt.Println("Made contact")
	network := NewNetwork()
	fmt.Println("Made network")

	payload := Payload{PING, "", nil, nil, network.table.me, contact, nil}
	fmt.Println("Make PING payload to compare")
	go network.SendPingMessage(&contact, send) //Without go channel in SendPing will block
	fmt.Println("Sent PING")
	recv := <-send
	fmt.Println("Received value from send channel")
	if recv.msg != payload.msg || recv.hash != payload.hash || recv.recipient != payload.recipient || recv.sender != payload.sender || !bytes.Equal(recv.data, payload.data) {
		t.Errorf("Wrong PING value sent")
	}

	fmt.Println("Make FINDNODE payload to compare")
	payload = Payload{FINDNODE, "", nil, nil, network.table.me, contact, nil}
	go network.SendFindContactMessage(&contact, &toContact, send)
	fmt.Println("Sent FINDNODE")
	recv = <-send
	fmt.Println("Received value from send channel")
	if recv.msg != payload.msg || recv.hash != payload.hash || recv.recipient != payload.recipient || recv.sender != payload.sender || !bytes.Equal(recv.data, payload.data) {
		t.Errorf("Wrong FINDNODE value sent")
	}

	fmt.Println("Make FINDDATA payload to compare")
	hash := "test"
	payload = Payload{FINDDATA, hash, nil, nil, network.table.me, contact, nil}
	go network.SendFindDataMessage(hash, &contact, send)
	fmt.Println("Sent FINDDATA")
	recv = <-send
	fmt.Println("Received value from send channel")
	if recv.msg != payload.msg || recv.hash != payload.hash || recv.recipient != payload.recipient || recv.sender != payload.sender || !bytes.Equal(recv.data, payload.data) {
		t.Errorf("Wrong FINDDATA value sent")
	}

	data := []byte("Test")
	createHash := sha1.Sum(data)
	hash = string(createHash[:])
	payload = Payload{STORE, hash, data, nil, network.table.me, contact, nil}
	go network.SendStoreMessage(data, &contact, send)
	recv = <-send
	if recv.msg != payload.msg || recv.hash != payload.hash || recv.recipient != payload.recipient || recv.sender != payload.sender || !bytes.Equal(recv.data, payload.data) {
		t.Errorf("Wrong STORE value sent")
	}
}
