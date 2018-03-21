// Entrypoint for the Distributed Diary Application
// This file can be run with 'go run distributeddiaryapp/app.go'
// Or do `cd distributeddiaryapp && go build && ./distributeddiaryapp`
// Or do `go install` then `distributeddiaryapp` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

package main

import (
	"fmt"
	"time"
	"os"
	"consensuslib"
	"distributeddiaryapp/cli"
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
	fmt.Printf("Alive: %v", isAlive)
	cli.InputLoop()
}

// TODO: add arg regex validation
func parseArgs(args []string) (serverAddr string, localAddr string, err error) {
	serverAddr = args[1]
	localAddr = args[2]
	return serverAddr, localAddr,  nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
