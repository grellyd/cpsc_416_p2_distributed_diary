// Entrypoint for the Distributed Diary Application
// This file can be run with 'go run distributeddiaryapp/app.go'
// Or do `cd distributeddiaryapp && go build && ./distributeddiaryapp`
// Or do `go install` then `distributeddiaryapp` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// USAGE: go run app.go SERVER IP:PORT [LOCAL PORT]
// Go Run Example (Dev): `go run distributeddiaryapp/app.go 127.0.0.1:12345 8080` -- To run on 127.0.0.1:8080
// Go Run Example (Prod): `go run distributeddiaryapp/app.go 127.0.0.1:12345` -- To run on machine's outbound IP
// Installed Run example: `distributeddiaryapp 127.0.0.1:12345 8080`

package main

import (
	"consensuslib"
	"distributeddiaryapp/cli"
	"fmt"
	"os"
	"time"
	"strconv"
)

func main() {
	client := setup()

	// BEGIN Test Code
	// (change in here to run when every app is started)

	/*
	isAlive, err := client.IsAlive()
	checkError(err)
	fmt.Printf("Alive: %v\n", isAlive)

	value, err := client.Read()
	checkError(err)
	fmt.Printf("Reading: '%s'\n", value)

	err = client.Write("Hello")
	checkError(err)

	value, err = client.Read()
	checkError(err)
	fmt.Printf("Reading: '%s'\n", value)
	*/

	// END Test Code

	serveCli(client)
}

func setup() *consensuslib.Client {
	serverAddr := ""
	localPort := 0
	// Validate arguments
	if len(os.Args[1:]) == 1 {
		// Local port not included; we're using a public IP so set to -1
		serverAddr = os.Args[1]
		localPort = -1
	} else if len(os.Args[1:]) == 2 {
		// Local port included; try to parse it
		intPort, err := strconv.Atoi(os.Args[2])
		if err != nil {
			printCommandLineUsageAndExit()
		}
		serverAddr = os.Args[1]
		localPort = intPort
	} else {
		printCommandLineUsageAndExit()
	}

	client, err := consensuslib.NewClient(localPort, 1*time.Millisecond)
	checkError(err)
	err = client.Connect(serverAddr)
	checkError(err)
	return client
}

func serveCli(client *consensuslib.Client) {
	for {
		command := cli.Run()
		switch command.Command {
		case cli.ALIVE:
			isAlive, err := client.IsAlive()
			checkError(err)
			fmt.Printf("[DD APP] Alive: %v\n", isAlive)
		case cli.EXIT:
			fmt.Println("[DD APP] Closing the Chamber of Secrets...")
			fmt.Println("[DD APP] Goodbye!")
			os.Exit(0)
		case cli.READ:
			value, err := client.Read()
			checkError(err)
			fmt.Printf("[DD APP] Reading: '%s'\n", value)
		case cli.WRITE:
			value := ""
			for i, s := range *command.Data {
				// add spaces
				if i != len(*command.Data) - 1 {
					value += s + " "
				} else {
					value += s
				}

			}
			err := client.Write(value)
			checkError(err)
		default:
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printCommandLineUsageAndExit() {
	fmt.Println("USAGE: go run app.go SERVERIP:PORT [LOCAL PORT]")
	os.Exit(1)
}
