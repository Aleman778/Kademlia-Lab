package main

type Payload struct {
	msg       RPCType
	hash      string
	data      []byte
	contacts  []Contact
	sender    Contact
	recipient Contact
	toContact *Contact
}

type RPCType int

const (
	PING RPCType = iota
	STORE
	FINDNODE
	FINDDATA
)
