// Entrypoint for the Distributed Diary Application
// This file can be run with 'go run distributeddiaryapp/app.go'
// Or do `cd distributeddiaryapp && go build && ./distributeddiaryapp`
// Or do `go install` then `distributeddiaryapp` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// USAGE: go run app.go [SERVER IP:PORT] [LOCAL PORT]
// Go Run Example (Dev): `go run distributeddiaryapp/app.go 127.0.0.1:12345 8080` -- To run on 127.0.0.1:8080
// Go Run Example (Prod): `go run distributeddiaryapp/app.go 127.0.0.1:12345 -1` -- To run on machine's outbound IP
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
	// Validate arguments
	serverAddr, localPort, err := parseArgs(os.Args)
	checkError(err)

	intPort, err := strconv.Atoi(localPort)
	if err != nil {
		printCommandLineUsage()
		os.Exit(1)
	}

	client, err := consensuslib.NewClient(intPort, 1*time.Millisecond)
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
			fmt.Printf("Alive: %v\n", isAlive)
		case cli.EXIT:
			fmt.Println("Closing the Chamber of Secrets...")
			fmt.Println("Goodbye!")
			os.Exit(0)
		case cli.READ:
			value, err := client.Read()
			checkError(err)
			fmt.Printf("Reading: '%s'\n", value)
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

// TODO: add arg regex validation
func parseArgs(args []string) (serverAddr string, localPort string, err error) {
	serverAddr = args[1]
	localPort = args[2]
	return serverAddr, localPort, nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printCommandLineUsage() {
	fmt.Println("USAGE: go run app.go [SERVER IP:PORT] [LOCAL PORT]")
}
