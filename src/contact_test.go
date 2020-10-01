package main

import (
    "testing"
)


func TestString(t *testing.T) {
    contact := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8080")
    if contact.String() != "contact(\"ffffffff00000000000000000000000000000000\", \"localhost:8080\")" {
        t.Errorf("Incorrect string formatting: %s", contact.String())
    }
}


func TestPopCandidates(t *testing.T) {
	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")

    var contacts []Contact
    contacts = append(contacts, contact1)
    contacts = append(contacts, contact2)
    contacts = append(contacts, contact3)

    contacts2, popped := PopCandidate(contacts)
    if popped != contact1 {
        t.Error()
    }

    if len(contacts2) != 2 {
        t.Error()
    }
}


func TestInCandidates(t *testing.T) {
	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")

    var contacts []Contact
    contacts = append(contacts, contact1)
    contacts = append(contacts, contact2)

    if !InCandidates(contacts, contact1) {
        t.Error()
    }
    
    if InCandidates(contacts, contact3) {
        t.Error()
    }
}


