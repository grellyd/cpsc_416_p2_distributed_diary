// Entrypoint for the Distributed Diary Server
// This file can be run with 'go run distributeddiaryserver/server.go'
// Or do `cd distributeddiaryserver && go build && ./distributeddiaryserver`
// Or do `go install` then `distributeddiaryserver` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// USAGE: go run server.go PORT [LOCAL]
// Go Run Example: `go run distributeddiaryserver/server.go 12345 LOCAL` -- To run server on 127.0.0.1:12345
// Go Run Example: `go run distributeddiaryserver/server.go 12345` -- To run server on the outbound IP address, on port 12345
// Installed Run example: `distributeddiaryserver 127.0.0.1:12345`

package main

import (
	"consensuslib"
	"fmt"
	"os"
	"strconv"
)

func main() {
	addr := ""

	// Validate arguments
	if len(os.Args[1:]) == 1 {
		// LOCAL arg not included; use public IP
		intPort, err := strconv.Atoi(os.Args[1])
		if err != nil {
			printCommandLineUsageAndExit()
		}
		addr = fmt.Sprintf(":%d", intPort)
	} else if len(os.Args[1:]) == 2 {
		// LOCAL included; use private IP
		intPort, err := strconv.Atoi(os.Args[1])
		if err != nil {
			printCommandLineUsageAndExit()
		}
		addr = fmt.Sprintf("127.0.0.1:%d", intPort)
	} else {
		printCommandLineUsageAndExit()
	}

	fmt.Printf("[DD SERVER] Calling consensuslib.NewServer with address %s\n", addr)
	server, err := consensuslib.NewServer(addr)
	checkError(err)
	err = server.Serve()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printCommandLineUsageAndExit() {
	fmt.Println("USAGE: go run server.go PORT [LOCAL?]")
	os.Exit(1)
}
