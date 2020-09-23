package main

import (
	"fmt"
	"os"
	"flag"
	"strings"
	"math/rand"
	"time"
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
    rand.Seed(time.Now().UTC().UnixNano())

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
	    data := strings.Join(os.Args[argIndex:], " ")
	    payload := Payload{data, []byte(data), nil}
            SendMessage(CliPut, payload);
            break;
        case "get":
            if len(os.Args) != argIndex + 1 || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia get\" require exactly 1 argument\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", GET_HELP_STRING);
                return;
            }

            hash := os.Args[argIndex];
	    payload := Payload{string(hash), nil, nil}
            fmt.Printf("Get data from hash: %s", hash);
            SendMessage(CliGet, payload);
            break;

        case "join":
            if len(os.Args) != argIndex + 1 || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia join\" requires exactly 1 arguments\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", JOIN_HELP_STRING);
                return;
            }
            address := os.Args[argIndex];
            fmt.Printf("Join node with address: %s\n", address);
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
	    payload := Payload{"", nil, nil}
            SendMessage(CliExit, payload);
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


func SendMessage(rpcType RPCType, payload Payload) {
	rpcMsg := RPCMessage{
		Type: rpcType,
		IsNode: false,
		Sender: NewContact(NewRandomKademliaID(), "client"),
		Payload: payload}

	conn := rpcMsg.SendTo("localhost:8080")
	defer conn.Close()


	response, _, err := GetRPCMessage(conn, 0)
	checkError(err)

	if response.Type == rpcType {
		switch (rpcType) {
		case CliPut:
		    fmt.Println("Data has been stored.");
		    break;
		case CliGet:
		    fmt.Printf("Data retrieved:\n%s\n", string(response.Payload.Data));
		    break;
		case CliExit:
		    fmt.Println("Node has been terminated");
		    break;
		}
	} else {
		fmt.Println("Failed to contact local server, start server using \"kademlia serve\"");
	}
}


func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

