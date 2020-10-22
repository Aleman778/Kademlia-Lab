package main

import (
	"testing"
    "strings"
    "os"
    "fmt"
)

func TestVersionFlag(t *testing.T) {
    GetRpcMessageFromCLI("kademlia --version", false)
    if !*showVersion {
        t.Error("expected flag to be set")
    }

    GetRpcMessageFromCLI("kademlia -v", false)
    if !*showV {
        t.Error("expected flag to be set")
    }

    GetRpcMessageFromCLI("kademlia --help", false)
    if !*showHelp {
        t.Error("expected flag to be set")
    }

    rpcMsg := GetRpcMessageFromCLI("kademlia put hello world", true)
    if rpcMsg.Type != CliPut {
        t.Error("unexpected rpc type")
    }

    if string(rpcMsg.Payload.Data) != "hello world" {
        t.Error("incorrect rpc data")
    }

    GetRpcMessageFromCLI("kademlia put", false)
    GetRpcMessageFromCLI("kademlia put Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.", false)

    rpcMsg = GetRpcMessageFromCLI("kademlia get 2aae6c35c94fcfb415dbe95f408b9ce91ee846ed", true)
    if rpcMsg.Type != CliGet {
        t.Error("unexpected rpc type")
    }

    if rpcMsg.Payload.Hash != "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed" {
        t.Error("incorrect rpc data")
    }
    
    GetRpcMessageFromCLI("kademlia get", false)
    GetRpcMessageFromCLI("kademlia get invalidhash", false)
    GetRpcMessageFromCLI("kademlia join", false)
    GetRpcMessageFromCLI("kademlia serve hello world", false)
    GetRpcMessageFromCLI("kademlia exit hello world", false)
}

func GetRpcMessageFromCLI(args string, expectResponse bool) RPCMessage {
	// Setup go routine for getting rpc messages
	getRpcCh := make(chan GetRPCConfig)
	defer close(getRpcCh)

	// Setup go routine for sending rpc messages
	sendToCh := make(chan SendToStruct)
	defer close(sendToCh)

    os.Args = strings.Split(args, " ")
    if expectResponse {
        go RunCLI(getRpcCh, sendToCh)
        data := <-sendToCh
        data.writeCh <- nil
        fmt.Println(data)
        value := <-getRpcCh
        fmt.Println(value)
        value.writeCh <- GetRPCData{data.rpcMsg, nil, nil}
        return data.rpcMsg
    } else {
        RunCLI(getRpcCh, sendToCh)
    }
    
    return RPCMessage{}
}
