package main

import (
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
)

const k = 5

type Network struct {
	table      *RoutingTable
	closest    *Contact
	candidates *ContactCandidates
	storage    *Storage
	mutex      sync.RWMutex
}

func NewNetwork() *Network {
	return &Network{NewRoutingTable(NewContact(NewRandomKademliaID(), "localhost:8000")), nil, nil, NewStorage(), sync.RWMutex{}}
}

func (network *Network) receive(send chan *Payload, conn net.Conn) {
	for {
		var recv Payload
		decoder := gob.NewDecoder(conn)
		decoder.Decode(&recv)
		switch recv.msg {
		case PING:
			go network.SendPingMessage(&recv.sender, send)
		case STORE:
			go network.storage.Store(recv.hash, recv.data)
		case FINDDATA:
			go func(payload Payload, send chan *Payload) {
				if payload.data == nil && payload.contacts == nil { //for someone else
					data, ok := network.storage.Load(recv.hash)
					if ok {
						send <- &Payload{FINDDATA, recv.hash, data, nil, network.table.me, recv.sender, nil}
					} else {
						hashID := NewKademliaID(recv.hash)
						closetContacts := network.table.FindClosestContacts(hashID, k)
						send <- &Payload{FINDDATA, "", nil, closetContacts, network.table.me, recv.sender, nil}
					}
				} else { //for me
					//
				}
			}(recv, send)
		case FINDNODE:
			go func(payload Payload, send chan *Payload) {
				if payload.toContact == nil { //for me
					//
				} else { //for someone else
					closetContacts := network.table.FindClosestContacts(recv.toContact.ID, k)
					send <- &Payload{FINDDATA, "", nil, closetContacts, network.table.me, recv.sender, nil}
				}
			}(recv, send)
		}
	}
}

func send(send chan *Payload) {
	for {
		payload := <-send
		remote := net.UDPAddr{IP: net.ParseIP(payload.recipient.Address), Port: 8080}
		conn, err := net.DialUDP("udp", nil, &remote)
		if err != nil {
			fmt.Printf("Error in send")
		}
		encoder := gob.NewEncoder(conn)
		encoder.Encode(*payload)
		conn.Close()
	}
}

func (network *Network) SendPingMessage(contact *Contact, send chan *Payload) {
	send <- &Payload{PING, "", nil, nil, network.table.me, *contact, nil}
}

func (network *Network) SendFindContactMessage(contact *Contact, toContact *Contact, send chan *Payload) {
	send <- &Payload{FINDNODE, "", nil, nil, network.table.me, *contact, toContact}
}

func (network *Network) SendFindDataMessage(hash string, contact *Contact, send chan *Payload) {
	send <- &Payload{FINDDATA, hash, nil, nil, network.table.me, *contact, nil}
}

func (network *Network) SendStoreMessage(data []byte, contact *Contact, send chan *Payload) {
	createHash := sha1.Sum(data)
	hash := string(createHash[:])
	send <- &Payload{STORE, hash, data, nil, network.table.me, *contact, nil}
}
