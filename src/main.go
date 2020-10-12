package main

func main() {
	//kad := NewKademlia("127.0.0.1:8888")
	//go kad.network.recv("127.0.0.1", 8888)

	/*for {
		fmt.Println("Enter input (put, get, exit): ")
		var input string
		fmt.Scanln(&input)
		splitInput := strings.Split(input, "")
		switch splitInput[0] {
		case "put":
			res, err := kad.Store([]byte(splitInput[1]))
			if err != nil {
				fmt.Println("Could not store data")
			} else {
				fmt.Println("Data stored: ", res)
			}
		case "get":
			contacts, data := kad.LookupData(splitInput[1])
			if data != nil {
				fmt.Println("Received data: ", data)
				fmt.Println("Received from node: ", contacts[0])
			} else {
				fmt.Println("Could not get data")
			}
		case "exit":
			break
		default:
			fmt.Println("Please enter get, put or exit")
		}
	}*/
}
