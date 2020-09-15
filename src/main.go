package main

import (
	"fmt"
	"net"
	"os"
    "flag"
    "strings"
)

const VERSION_STRING string    = "Kademlia version 1.0.0"
const PUT_HELP_STRING string   = "Usage:  kademlia put DATA        Note: maximum of 255 characters allowed\n\n" +
                                 "Stores the data in the hash table and returns the hash key used for get command"
const GET_HELP_STRING string   = "Usage:  kademlia get HASH\n\n" +
                                 "Retrieve previously stored data using the hash key"
const JOIN_HELP_STRING string  = "Usage:  kademlia join ADDRESS\n\n" +
                                 "Join another node with the given address"
const SERVE_HELP_STRING string = "Usage:  kademlia serve\n\n" +
                                 "Starts the kademlia server on this node"
const EXIT_HELP_STRING string  = "Usage:  kademlia kill\n\n" +
                                 "Kills the currently running kademlia node"


func main() {
    // Setup CLI arguments
    showVersion := flag.Bool("-version", false, "Print version number and quit");
    showV := flag.Bool("v", false, "Print version number and quit");
    showHelp := flag.Bool("-help", false, "Show help information");

    // Filter CLI command from args, for some reason golang flags fails if there are commands in the arguments.
    command := "";
    argIndex := 1;
    for i, arg := range os.Args[1:] {
        if !strings.HasPrefix(arg, "-") {
            command = os.Args[i + 1];
            os.Args = append(os.Args[:i + 1], os.Args[i + 2:]...);
            break;
        }
    }
    
    // Parse flags
    flag.Parse();
    if (*showVersion || *showV) {
        fmt.Printf(VERSION_STRING);
        return;
    }
    
    // Run CLI commands
    if command != "" {
        // flag.Parse();
        switch command {
        case "put":
            if len(os.Args) <= argIndex || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia put\" require at least 1 argument\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", PUT_HELP_STRING);
                return;
            }
            data := strings.Join(os.Args[argIndex:], " ");
            fmt.Printf("Put data: %s", data);
            // StoreData(data);
            break;
        case "get":
            if len(os.Args) != argIndex + 1 || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia get\" require exactly 1 argument\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", GET_HELP_STRING);
                return;
            }
            hash := os.Args[argIndex];
            fmt.Printf("Get data from hash: %s", hash);
            // GetData(hash)
            break;

        case "join":
            if len(os.Args) != argIndex + 1 || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia join\" requires exactly 1 arguments\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", JOIN_HELP_STRING);
                return;
            }
            address := os.Args[argIndex];
            fmt.Printf("Join node with address: %s", address);
            JoinNetwork(address);

        case "serve":
            if len(os.Args) > argIndex || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia serve\" has no arguments\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", SERVE_HELP_STRING);
                return;
            }
            
            InitServer();
            break;

        case "exit":
            if len(os.Args) > argIndex || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia exit\" has no arguments\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", EXIT_HELP_STRING);
                return;
            }
            client("localhost:8080", ExitNode);
            break;
        default:
            fmt.Printf("\nInvalid command '%s' was found\n", command);
            *showHelp = true;
            break;
        }

    }

    if command == "" || *showHelp {
        fmt.Print("\n");
        fmt.Print("Usage:  kademlia [OPTIONS] COMMAND\n");
        fmt.Print("\n");
        fmt.Print("Distributed hash table for decentralized peer-to-peer computer networks\n");
        fmt.Print("\n");
        fmt.Print("Options:\n");
        fmt.Print("  -v, --version        Print version number and quit\n");
        fmt.Print("      --help           Show help information\n");
        fmt.Print("\n");
        fmt.Print("Commands:\n");
        fmt.Print("  put       Store data in the hash table and returns the hash key used for get command\n");
        fmt.Print("  get       Retrieve previously stored data using the hash key\n");
        fmt.Print("  join      Join another node with the given address\n");
        fmt.Print("  serve     Starts the kademlia server on this node\n");
        fmt.Print("  exit      Kills the currently running kademlia node\n");
        fmt.Print("\n");
        fmt.Print("Run 'kademlia COMMAND --help' for more information on a command.\n");
    }
}




func client(service string, rpc RPCType) {
	rpcMsg := RPCMessage{
		Type: rpc,
		Me: NewContact(NewRandomKademliaID(), ""),
		Data: []byte(nil)}

	switch rpcMsg.Type {
	case Ping:
	case Store:
	case FindNode:
		rpcMsg.Data = EncodeKademliaID(*NewRandomKademliaID())
	case FindValue:
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)

	conn, err := net.DialUDP("udp4", nil, udpAddr)
	checkError(err)

	defer conn.Close()

	_, err = conn.Write(EncodeRPCMessage(rpcMsg))
	checkError(err)

	inputBytes := make([]byte, 1024)
	length, _, _ := conn.ReadFromUDP(inputBytes)

	var rrpcMsg RPCMessage
	DecodeRPCMessage(&rrpcMsg, inputBytes[:length])
	fmt.Println(rrpcMsg.String())

	switch rrpcMsg.Type {
	case Ping:
	case Store:
	case FindNode:
		var contacts  []Contact
		DecodeContacts(&contacts, rrpcMsg.Data)
		fmt.Println("Contacts Decoded: ", contacts)
	case FindValue:
    case ExitNode:
        fmt.Println("Node has been terminated");
	}

	fmt.Println("")
}


func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

