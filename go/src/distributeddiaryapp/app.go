// Entrypoint for the Distributed Diary Application
// This file can be run with 'go run distributeddiaryapp/app.go'
// Or do `cd distributeddiaryapp && go build && ./distributeddiaryapp`
// Or do `go install` then `distributeddiaryapp` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// Go Run Example: `go run distributeddiaryapp/app.go 127.0.0.1:12345 127.0.0.1:0`
// Installed Run example: `distributeddiaryapp 127.0.0.1:12345 127.0.0.1:0`

package main

import (
	"consensuslib"
	"distributeddiaryapp/cli"
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("testing app")
	serverAddr, localAddr, err := parseArgs(os.Args)
	checkError(err)
	client, err := consensuslib.NewClient(localAddr, 1*time.Millisecond)
	checkError(err)
	err = client.Connect(serverAddr)
	checkError(err)
	isAlive, err := client.IsAlive()
	checkError(err)
	value, err := client.Read()
	fmt.Printf("Alive: %v\n", isAlive)
	checkError(err)
	fmt.Printf("Reading: '%s'\n", value)
	err = client.Write("Hello")
	checkError(err)
	value, err = client.Read()
	checkError(err)
	fmt.Printf("Reading: '%s'\n", value)
	serveCli(client)

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
			for _, s := range *command.Data {
				// TODO: removes spaces
				value += s
			}
			err := client.Write(value)
			checkError(err)
		default:
		}
	}
}

// TODO: add arg regex validation
func parseArgs(args []string) (serverAddr string, localAddr string, err error) {
	serverAddr = args[1]
	localAddr = args[2]
	return serverAddr, localAddr, nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
