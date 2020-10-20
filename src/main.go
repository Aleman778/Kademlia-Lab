package main

import (
	"fmt"
	"os"
    "io"
	"flag"
	"strings"
	"math/rand"
    "crypto/sha1"
    "encoding/hex"
	"time"
)

const VERSION_STRING string     = "Kademlia version 1.0.0"
const PUT_HELP_STRING string    = "Usage:  kademlia put DATA        Note: maximum of 255 characters allowed\n\n" +
                                  "Stores the data in the hash table and returns the hash key used for get command"
const GET_HELP_STRING string    = "Usage:  kademlia get HASH\n\n" +
                                  "Retrieve previously stored data using the hash key"
const FORGET_HELP_STRING string = "Usage:  kademlia forget HASH\n\n" +
                                  "Forgets previously stored data using the hash key"
const JOIN_HELP_STRING string   = "Usage:  kademlia join ADDRESS\n\n" +
                                  "Join another node with the given address"
const SERVE_HELP_STRING string  = "Usage:  kademlia serve\n\n" +
                                  "Starts the kademlia server on this node"
const EXIT_HELP_STRING string   = "Usage:  kademlia kill\n\n" +
                                  "Kills the currently running kademlia node"

func main() {
    rand.Seed(time.Now().UTC().UnixNano())

	// Setup go routine for getting rpc messages
	getRpcCh := make(chan GetRPCConfig)
	defer close(getRpcCh)
	go GetRPCMessageStarter(getRpcCh)

	// Setup go routine for sending rpc messages
	sendToCh := make(chan SendToStruct)
	defer close(sendToCh)
	go SendToStarter(sendToCh)

    // Setup CLI arguments
    showVersion := flag.Bool("version", false, "Print version number and quit");
    showV := flag.Bool("v", false, "Print version number and quit");
    showHelp := flag.Bool("help", false, "Show help information");
    objectTTL := flag.Int64("ttl", maxExpire, "Objects time to live");

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
    for _, arg := range os.Args[1:] {
        if strings.HasPrefix(arg, "-") {
            argIndex += 1;
        }
    }

    // Parse flags
    flag.Parse();
    if (*showVersion || *showV) {
        fmt.Printf(VERSION_STRING);
        return;
    }

    if (*objectTTL <= 5) {
        fmt.Printf("error: time to live is required to be larger than 5 (got %d)\n", *objectTTL);
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
            if len(data) > 255 {
                fmt.Printf("\nCan't send data that is longer than 255 characters (got %d characters).\n", len(data))
                os.Exit(1)
            }
            hash := sha1.Sum([]byte(data))
            hash_string := hex.EncodeToString(hash[:])
            payload := Payload{string(hash_string), []byte(data), *objectTTL, nil}
            SendMessage(getRpcCh, sendToCh, CliPut, payload, os.Stdout);
            break;

        case "get":
            if len(os.Args) != argIndex + 1 || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia get\" require exactly 1 argument\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", GET_HELP_STRING);
                return;
            }

            hash := os.Args[argIndex];
            decoded, err := hex.DecodeString(hash)
            if err != nil {
                fmt.Print(err)
            }

            if len(decoded) != IDLength {
                fmt.Printf("error: expected 160-bit hash in hexadecimal format\n");
                return;
            }

            payload := Payload{string(hash), nil, *objectTTL, nil}
            SendMessage(getRpcCh, sendToCh, CliGet, payload, os.Stdout);
            break;

        case "forget":
            if len(os.Args) != argIndex + 1 || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia forget\" require exactly 1 argument\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", FORGET_HELP_STRING);
                return;
            }
            
            hash := os.Args[argIndex];
            decoded, err := hex.DecodeString(hash)
            if err != nil {
                fmt.Print(err)
            }

            if len(decoded) != IDLength {
                fmt.Printf("error: expected 160-bit hash in hexadecimal format\n");
                return;
            }

            payload := Payload{string(hash), nil, *objectTTL, nil}
            SendMessage(getRpcCh, sendToCh, CliForget, payload, os.Stdout);

        case "join":
            if len(os.Args) != argIndex + 1 || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia join\" requires exactly 1 arguments\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", JOIN_HELP_STRING);
                return;
            }
            address := os.Args[argIndex];
            fmt.Printf("\nJoin node with address: %s\n", address);
            JoinNetwork(address);

        case "serve":
            if len(os.Args) > argIndex || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia serve\" has no arguments\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", SERVE_HELP_STRING);
                return;
            }

            fmt.Print("\n");
            InitServer();
            break;

        case "exit":
            if len(os.Args) > argIndex || *showHelp {
                if !*showHelp { fmt.Printf("\"kademlia exit\" has no arguments\n"); }
                fmt.Print("\n");
                fmt.Printf("%s\n", EXIT_HELP_STRING);
                return;
            }
            payload := Payload{"", nil, *objectTTL, nil}
            SendMessage(getRpcCh, sendToCh, CliExit, payload, os.Stdout);
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
        fmt.Print("      --ttl            Objects time to live\n");
        fmt.Print("\n");
        fmt.Print("Commands:\n");
        fmt.Print("  put       Store data in the hash table and returns the hash key used for get command\n");
        fmt.Print("  get       Retrieve previously stored data using the hash key\n");
        fmt.Print("  forget    Froget previously stored data using the hash key\n");
        fmt.Print("  join      Join another node with the given address\n");
        fmt.Print("  serve     Starts the kademlia server on this node\n");
        fmt.Print("  exit      Kills the currently running kademlia node\n");
        fmt.Print("\n");
        fmt.Print("Run 'kademlia COMMAND --help' for more information on a command.\n");
    }
}

func SendMessage(getRpcCh chan<- GetRPCConfig, sendToCh chan<- SendToStruct, rpcType RPCType, payload Payload, w io.Writer) {
	rpcMsg := RPCMessage{
		Type: rpcType,
		IsNode: false,
		Sender: NewContact(NewRandomKademliaID(), "client"),
		Payload: payload}

	conn := rpcMsg.SendTo(sendToCh, "localhost:8080", false)
    fmt.Fprintln(w, "\nSending request to local kademlia server...");
	defer conn.Close()

	readCh := make(chan GetRPCData)
	getRpcCh <- GetRPCConfig{readCh, conn, 30, false}
	data := <-readCh
	response := data.rpcMsg
	err := data.err

	if err != nil {
        fmt.Fprintln(w, "Local server is not responding, start the server using \"kademlia serve\"");
        os.Exit(1)
    }

	if response.Type == rpcType {
		switch (rpcType) {
		case CliPut:
		    fmt.Fprintf(w, "Data has been stored. Copy your hash:\n%s\n", response.Payload.Hash);
		    break;
		case CliGet:
		    fmt.Fprintf(w, "Data retrieved:\n%s\n", string(response.Payload.Data));
		    break;
		case CliForget:
		    fmt.Fprintf(w, "Data stored with the given hash has been forgotten!\n");
		    break;
		case CliExit:
		    fmt.Fprintln(w, "Node has been terminated");
		    break;
		}
	} else {
		fmt.Fprintln(w, "Failed to contact local server, start the server using \"kademlia serve\"");
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

