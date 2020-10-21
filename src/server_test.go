package main

import (
	"fmt"
	"strings"
	"testing"
	"strconv"
	"crypto/sha1"
	"encoding/hex"
	"sort"
)

type TestError struct {
	msg string
}

func (e *TestError) Error() string {
    return fmt.Sprintf(e.msg)
}

func TestCreateServer(t *testing.T) {
	server := CreateServer()
	contact := server.table.GetMe()
	if !strings.HasSuffix(contact.Address, PORT) {
		t.Errorf("Expected the address to end with %s got %s", PORT, contact.Address)
	}
	destroyServer(&server)
}

func TestBootStrapNode(t *testing.T) {
	server, getRpcCh, sendToCh, _ := createServer()
	address := "localhost:1231"

	me := server.table.GetMe()
	meId := me.ID
	me.CalcDistance(meId)

	contact1 := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")
	contact1.CalcDistance(meId)
	contact2 := NewContact(NewKademliaID("0000000000000000000000000000000000000111"), "localhost:8002")
	contact2.CalcDistance(meId)

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindNode,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"",nil, maxExpire, []Contact{contact1},},
		},
		addr: nil,
		err: nil,
	}

	go server.BootstrapNode(address)

	fakeSendToResponder(sendToCh)
	fakeGetRPCMessage(getRpcCh, response)

	response = GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindNode,
			IsNode: true,
			Sender: contact1,
			Payload: Payload{"",nil, maxExpire, []Contact{contact2, contact1},},
		},
		addr: nil,
		err: nil,
	}

	fakeSendToResponder(sendToCh)
	fakeGetRPCMessage(getRpcCh, response)

	test_err := TestError{msg: "Test error"}
	response = GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindNode,
			IsNode: true,
			Sender: contact2,
			Payload: Payload{"",nil, maxExpire, nil,},
		},
		addr: nil,
		err: &test_err,
	}

	fakeSendToResponder(sendToCh)
	fakeGetRPCMessage(getRpcCh, response)

	expContacts := []Contact{contact1, me}
	sort.Sort(ByDistance(expContacts))

	contacts := server.table.FindClosestContacts(meId, k)

	if len(contacts) != len(expContacts) {
		t.Error("Expected ",  len(expContacts), " contacts got ", len(contacts))
	} else {
		for i := 0; i < len(expContacts); i++ {
			if !contacts[i].ID.Equals(expContacts[i].ID) {
				t.Error("Expected contacts ", expContacts, " got ", contacts)
			}
		}
	}

	destroyServer(server)
}

func TestValueLookup(t *testing.T) {
	server, _, sendToCh, _ := createServer()

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	go func(){
		fakeSendToResponder(sendToCh)
	}()

	//go server.ValueLookup(hash_string, maxExpire)
	result := server.ValueLookup(hash_string, maxExpire)

	verifyPayload(t, result, "", nil, nil)

	destroyServer(server)
}

func TestValueLookupSender(t *testing.T) {
	server, getRpcCh, sendToCh, _ := createServer()

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	contactsCh := make(chan LookupResponse)
	defer close(contactsCh)

	contactCh := make(chan Contact)

	intermediateCh := make(chan Payload)
	defer close(intermediateCh)

	go server.ValueLookupSender(hash_string, contactsCh, contactCh, intermediateCh)

	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")

	contactCh <- contact

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{hash_string, []byte(data), maxExpire, []Contact{},},
		},
		addr: nil,
		err: nil,
	}

	fakeSendToResponder(sendToCh)
	fakeGetRPCMessage(getRpcCh, response)

	result := <-intermediateCh

	verifyPayload(t, result, hash_string, []byte(data), []Contact{})

	lookupResponse := <-contactsCh
	if len(lookupResponse.contacts) != 0 {
		t.Error("Expected 0 contacts got ", len(lookupResponse.contacts))
	}
	if !lookupResponse.hasValue {
		t.Error("Expected hasValue true got ", lookupResponse.hasValue)
	}
	if !lookupResponse.contact.ID.Equals(contact.ID) {
		t.Error("Expected contact ", contact, " got ", lookupResponse.contact)
	}

	contactCh <- contact

	test_err := TestError{msg: "Test error"}
	response = GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"", nil, maxExpire, []Contact{},},
		},
		addr: nil,
		err: &test_err,
	}

	fakeSendToResponder(sendToCh)
	fakeGetRPCMessage(getRpcCh, response)


	lookupResponse = <-contactsCh
	if len(lookupResponse.contacts) != 0 {
		t.Error("Expected 0 contacts got ", len(lookupResponse.contacts))
	}
	if !lookupResponse.hasValue {
		t.Error("Expected hasValue true got ", lookupResponse.hasValue)
	}
	if !lookupResponse.contact.ID.Equals(contact.ID) {
		t.Error("Expected contact ", contact, " got ", lookupResponse.contact)
	}

	close(contactCh)
	destroyServer(server)
}

func TestValueLookupSenderIsDoneFalseCase(t *testing.T) {
	server, _, _, _ := createServer()

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	contactsCh := make(chan LookupResponse)
	defer close(contactsCh)

	contactCh := make(chan Contact)

	intermediateCh := make(chan Payload)
	defer close(intermediateCh)

	go server.ValueLookupSender(hash_string, contactsCh, contactCh, intermediateCh)

	close(contactCh)

	result := <-intermediateCh

	verifyPayload(t, result, "", nil, nil)

	destroyServer(server)
}

func TestSendFindDataMessage(t *testing.T) {
	server, getRpcCh, sendToCh, _ := createServer()
	address := "localhost:1231"

	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")
	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"",nil, maxExpire, []Contact{contact},},
		},
		addr: nil,
		err: nil,
	}

	go fakeSendToResponder(sendToCh)
	go fakeGetRPCMessage(getRpcCh, response)

	payload, err := server.SendFindDataMessage(address, hash_string)

	if len(payload.Contacts) != 1 {
		t.Errorf("Expected 1 got %d contacts", len(payload.Contacts))
	} else {
		if !payload.Contacts[0].ID.Equals(contact.ID) {
			t.Error("Expected contact ", contact, " got ", payload.Contacts[0])
		}
	}
	if err != nil {
		t.Error("Expected no error")
	}

	destroyServer(server)
}

func TestSendFindDataMessageError(t *testing.T) {
	server, getRpcCh, sendToCh, _ := createServer()
	address := "localhost:1231"

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	test_err := TestError{msg: "Test error"}
	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"",nil, maxExpire, []Contact{},},
		},
		addr: nil,
		err: &test_err,
	}

	go fakeSendToResponder(sendToCh)
	go fakeGetRPCMessage(getRpcCh, response)

	payload, err := server.SendFindDataMessage(address, hash_string)

	if len(payload.Contacts) != 0 {
		t.Errorf("Expected 0 got %d contacts", len(payload.Contacts))
	}

	if err == nil {
		t.Error("Expected error")
	}

	destroyServer(server)
}

func TestSendFindContactMessage(t *testing.T) {
	server, getRpcCh, sendToCh, _ := createServer()
	address := "localhost:1231"
	id := NewKademliaID("0000000000000000000000000000000000000211")

	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindNode,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"",nil, maxExpire, []Contact{contact},},
		},
		addr: nil,
		err: nil,
	}

	go fakeSendToResponder(sendToCh)
	go fakeGetRPCMessage(getRpcCh, response)

	contacts, err := server.SendFindContactMessage(address, *id, maxExpire)

	if len(contacts) != 1 {
		t.Errorf("Expected 1 got %d contacts", len(contacts))
	} else {
		if !contacts[0].ID.Equals(contact.ID) {
			t.Error("Expected contact ", contact, " got ", contacts[0])
		}
	}
	if err != nil {
		t.Error("Expected no error")
	}

	destroyServer(server)
}

func TestSendFindContactMessageError(t *testing.T) {
	server, getRpcCh, sendToCh, _ := createServer()
	address := "localhost:1231"
	id := NewKademliaID("0000000000000000000000000000000000000211")

	test_err := TestError{msg: "Test error"}
	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindNode,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"",nil, maxExpire, nil,},
		},
		addr: nil,
		err: &test_err,
	}

	go fakeSendToResponder(sendToCh)
	go fakeGetRPCMessage(getRpcCh, response)

	contacts, err := server.SendFindContactMessage(address, *id, maxExpire)

	if len(contacts) != 0 {
		t.Errorf("Expected 0 got %d contacts", len(contacts))
	}
	if err == nil {
		t.Error("Expected error")
	}

	destroyServer(server)
}

func TestHandleClientError(t *testing.T) {
	server, getRpcCh, _, _ := createServer()

	test_err := TestError{msg: "Test error"}
	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: Ping,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"",nil, maxExpire, nil,},
		},
		addr: nil,
		err: &test_err,
	}

	go fakeGetRPCMessage(getRpcCh, response)

	server.HandleClient(nil)

	destroyServer(server)
}

func TestHandlePingMessage(t *testing.T) {
	server, getRpcCh, _, sendResponseCh := createServer()

	go server.HandleClient(nil)
	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")
	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: Ping,
			IsNode: true,
			Sender: contact,
			Payload: Payload{"",nil, maxExpire, nil,},
		},
		addr: nil,
		err: nil,
	}
	go fakeGetRPCMessage(getRpcCh, response)

	result := <-sendResponseCh

	verifyRPCMessage(t, result.rpcMsg, Ping, true, server.table.GetMe())
	verifyPayload(t, result.rpcMsg.Payload, "", nil, nil)

	destroyServer(server)
}

func TestHandleStoreMessage(t *testing.T) {
	server, getRpcCh, _, sendResponseCh := createServer()
	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")
	server.AddContact(contact)

	go server.HandleClient(nil)

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])
	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: Store,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{hash_string, []byte(data), maxExpire, nil,},
		},
		addr: nil,
		err: nil,
	}

	go fakeGetRPCMessage(getRpcCh, response)

	result := <-sendResponseCh

	verifyRPCMessage(t, result.rpcMsg, Store, true, server.table.GetMe())
	verifyPayload(t, result.rpcMsg.Payload, "", nil, nil)

	if val, ok := server.storage.Load(hash_string, result.rpcMsg.Payload.TTL); ok {
		if string(val) != data {
			t.Errorf("Expected %s got %s", data, string(val))
		}
	} else {
		t.Errorf("Expected %s to be in storage", data)
	}

	destroyServer(server)
}


func TestHandleFindNodeMessage(t *testing.T) {
	server, getRpcCh, _, sendResponseCh := createServer()
	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")
	server.AddContact(contact)

	go server.HandleClient(nil)

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindNode,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"", nil, maxExpire, []Contact{contact},},
		},
		addr: nil,
		err: nil,
	}
	go fakeGetRPCMessage(getRpcCh, response)

	result := <-sendResponseCh
	var id KademliaID = *response.rpcMsg.Payload.Contacts[0].ID
        contacts := server.table.FindClosestContacts(&id, k)

	verifyRPCMessage(t, result.rpcMsg, FindNode, true, server.table.GetMe())
	verifyPayload(t, result.rpcMsg.Payload, "", nil, contacts)

	destroyServer(server)
}

func TestHandleFindValueMessage(t *testing.T) {
	server, getRpcCh, _, sendResponseCh := createServer()
	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")
	server.AddContact(contact)

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	server.storage.Store(hash_string, []byte(data), maxExpire)

	go server.HandleClient(nil)

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: FindValue,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{hash_string,nil, maxExpire, nil,},
		},
		addr: nil,
		err: nil,
	}
	go fakeGetRPCMessage(getRpcCh, response)

	result := <-sendResponseCh

	verifyRPCMessage(t, result.rpcMsg, FindValue, true, server.table.GetMe())
	verifyPayload(t, result.rpcMsg.Payload, hash_string, []byte(data), []Contact{})

	destroyServer(server)
}

func TestHandleFindValueMessageFail(t *testing.T) {
	server, _, _, sendResponseCh := createServer()
	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")
	server.AddContact(contact)

	hash := sha1.Sum([]byte("test"))
	hash_string := hex.EncodeToString(hash[:])
	rpcMsg := RPCMessage{
		Type: FindValue,
		IsNode: true,
		Sender: server.table.GetMe(),
		Payload: Payload{hash_string,nil, maxExpire, nil,},
	}

	go server.HandleFindValueMessage(&rpcMsg, nil, nil)

	result := <-sendResponseCh

	verifyRPCMessage(t, result.rpcMsg, FindValue, true, server.table.GetMe())
	verifyPayload(t, result.rpcMsg.Payload, "", nil, []Contact{contact})

	destroyServer(server)
}

func TestHandleCliPutMessage(t *testing.T) {
	server, getRpcCh, _, sendResponseCh := createServer()
	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	go server.HandleClient(nil)

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: CliPut,
			IsNode: false,
			Sender: contact,
			Payload: Payload{hash_string, []byte(data), maxExpire, nil,},
		},
		addr: nil,
		err: nil,
	}
	go fakeGetRPCMessage(getRpcCh, response)

	result := <-sendResponseCh

	verifyRPCMessage(t, result.rpcMsg, CliPut, true, server.table.GetMe())
	verifyPayload(t, result.rpcMsg.Payload, hash_string, nil, []Contact{})
}

func TestHandleCliGetMessage(t *testing.T) {
	server, getRpcCh, _, sendResponseCh := createServer()
	contact := NewContact(NewKademliaID("0000000000000000000000000000000000000211"), "localhost:8002")

	data := "test"
	hash := sha1.Sum([]byte(data))
	hash_string := hex.EncodeToString(hash[:])

	server.storage.Store(hash_string, []byte(data), maxExpire)

	go server.HandleClient(nil)

	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: CliGet,
			IsNode: false,
			Sender: contact,
			Payload: Payload{hash_string, nil, maxExpire, nil,},
		},
		addr: nil,
		err: nil,
	}
	go fakeGetRPCMessage(getRpcCh, response)

	result := <-sendResponseCh

	verifyRPCMessage(t, result.rpcMsg, CliGet, true, server.table.GetMe())
	verifyPayload(t, result.rpcMsg.Payload, hash_string, []byte(data), []Contact{})

	destroyServer(server)
}

func TestServerAddContact(t *testing.T) {
	server, getRpcCh, sendToCh, _ := createServer()

	go fakeSendToResponder(sendToCh)

	err := TestError{msg: "Test error"}
	response := GetRPCData{
		rpcMsg: RPCMessage{
			Type: Ping,
			IsNode: true,
			Sender: server.table.GetMe(),
			Payload: Payload{"",nil, maxExpire, nil,},
		},
		addr: nil,
		err: &err,
	}
	go fakeGetRPCMessage(getRpcCh, response)

	removed := NewContact(NewKademliaID("00000000000000000000000000000000000000" + strconv.Itoa(bucketSize - 1)), "localhost:8002")
	for i := 0; i <= bucketSize; i++ {
		var contact Contact
		if i > 9 {
			contact = NewContact(NewKademliaID("00000000000000000000000000000000000000" + strconv.Itoa(i)), "localhost:8002")
		} else {
			contact = NewContact(NewKademliaID("000000000000000000000000000000000000000" + strconv.Itoa(i)), "localhost:8002")
		}

		server.AddContact(contact)
	}

	contacts := server.table.FindClosestContacts(removed.ID, 1)
	if contacts[0].ID.Equals(removed.ID) {
		t.Errorf("Expected %s to not be in RoutingTable", removed.String())
	}

	destroyServer(server)
}


func createServer() (*Server, chan GetRPCConfig, chan SendToStruct, chan SendResponseStruct ) {
	me := NewContact(NewRandomKademliaID(), ":6060")
	getRpcCh := make(chan GetRPCConfig)
	sendToCh := make(chan SendToStruct)
	sendResponseCh := make(chan SendResponseStruct)
	return &Server{
		table: NewRoutingTable(me),
		storage: NewStorage(),

		getRpcCh: getRpcCh,

		sendToCh: sendToCh,
		sendResponseCh: sendResponseCh,
	}, getRpcCh, sendToCh, sendResponseCh
}

func destroyServer(server *Server) {
	close(server.getRpcCh)
	close(server.sendToCh)
	close(server.sendResponseCh)
}

func fakeGetRPCMessage(readCh <-chan GetRPCConfig, response GetRPCData) {
	value := <-readCh
	value.writeCh <- response
	close(value.writeCh)
}

func fakeSendToResponder(readCh <-chan SendToStruct) {
	value := <-readCh
	value.writeCh <- nil
	close(value.writeCh)
}

func fakeSendResponseResponder(readCh <-chan SendResponseStruct) {
	<-readCh
}

func verifyRPCMessage(t *testing.T, rpcMsg RPCMessage, expType RPCType, expIsNode bool, expSender Contact) {
	if rpcMsg.Type != expType {
		t.Errorf("Expected RPCMsg type %d got %d", expType, rpcMsg.Type)
	}

	if rpcMsg.IsNode != expIsNode {
		t.Errorf("Expected RPCMsg IsNode %t got %t", expIsNode, rpcMsg.IsNode)
	}

	if !rpcMsg.Sender.ID.Equals(expSender.ID) {
		t.Error("Expected RPCMsg Sender \n", expSender.String(), "\n got\n", rpcMsg.Sender.String())
	}
}

func verifyPayload(t *testing.T, payload Payload, expHash string, expData []byte, expContacts []Contact) {
	if payload.Hash != expHash {
		t.Errorf("Expected hash %s got %s", expHash, payload.Hash)
	}

	if len(payload.Data) != len(expData) {
		t.Errorf("Expected data %s got %s", expData, payload.Data)
	}

	for i := 0; i < len(expData); i++ {
		if payload.Data[i] != expData[i] {
			t.Errorf("Expected data %s got %s", expData, payload.Data)
		}
	}

	if len(payload.Contacts) != len(expContacts) {
		t.Error("Expected ",  len(expContacts), " contacts got ", len(payload.Contacts))
	}

	for i := 0; i < len(expContacts); i++ {
		if !payload.Contacts[i].ID.Equals(expContacts[i].ID) {
			t.Error("Expected contacts ", expContacts, " got ", payload.Contacts)
		}
	}
}

