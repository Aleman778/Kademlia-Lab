package main

import (
	"testing"
)

func TestRoutingTableAddContact(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))

	id :=NewKademliaID("1111111100000000000000000000000000000000")
	rt.AddContact(NewContact(id, "localhost:8002"))

	length := rt.buckets[0].Len()
	if length != 1 {
		t.Errorf("Expected bucket lengt 1 got %d", length)
	}
}

func TestFindClosestContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:7000"))

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")

	rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8001"))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(contact3)
	rt.AddContact(contact2)
	rt.AddContact(contact1)

	contacts := rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 20)
	length := len(contacts)
	if length != 6 {
		t.Errorf("Expected to find 6 contacts got %d", length)
	}

	contacts = rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 6)
	length = len(contacts)
	if length != 6 {
		t.Errorf("Expected to find 6 contacts got %d", length)
	}

	contacts = rt.FindClosestContacts(NewKademliaID("2111111400000000000000000000000000000000"), 3)
	length = len(contacts)
	if length != 3 {
		t.Errorf("Expected to find 3 contacts got %d", length)
	}

	if !contact1.ID.Equals(contacts[0].ID) {
		t.Error("Expected to find the closest contact \n", contact1.String(), "\n got\n", contacts[0].String())
	}
	if !contact2.ID.Equals(contacts[1].ID) {
		t.Error("Expected to find the closest contact \n", contact2.String(), "\n got\n", contacts[1].String())
	}
	if !contact3.ID.Equals(contacts[2].ID) {
		t.Error("Expected to find the closest contact \n", contact3.String(), "\n got\n", contacts[2].String())
	}

}

func TestGetBucketIndex(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"))

	firstId :=NewKademliaID("1111111100000000000000000000000000000000")
	rt.AddContact(NewContact(firstId, "localhost:8002"))

	lastId :=NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	rt.AddContact(NewContact(lastId, "localhost:8002"))

	index := rt.getBucketIndex(firstId)
	if index != 0 {
		t.Errorf("Expected bucket index 0 got %d", index)
	}

	index = rt.getBucketIndex(lastId)
	if index != 159 {
		t.Errorf("Expected bucket index 159 got %d", index)
	}
}


func TestGetMe(t *testing.T) {
	me := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	rt := NewRoutingTable(me)

	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002"))

	result := rt.GetMe()
	if !me.ID.Equals(result.ID) {
		t.Error("Expected Contact \n", me.String(), "\n got\n", result.String())
	}
}

