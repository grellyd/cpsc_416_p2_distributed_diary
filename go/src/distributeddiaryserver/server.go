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
	"fmt"
	"consensuslib"
	"filelogger"
	"os"
	"strconv"
	"regexp"
)

const (
	serverAddrDefault = "127.0.0.1:12345"
)

var logger *filelogger.Logger
var validAddr = regexp.MustCompile("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:[0-9]{1,5}")

func main() {
	var err error
	logger, err = filelogger.NewFileLogger("server", filelogger.NORMAL)
	checkError(err)
	logger.Info("Logger created")
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

	// addr, err := parseArgs(os.Args)
	// checkError(err)
	
	fmt.Printf("[DD SERVER] Calling consensuslib.NewServer with address %s\n", addr)
	server, err := consensuslib.NewServer(addr, logger)
	checkError(err)
	err = server.Serve()
	checkError(err)
}

func parseArgs(args []string) (serverAddr string, err error) {
	if len(args) > 1 {
		serverAddr = args[1] 
		if !validAddr.MatchString(serverAddr) {
			logger.Error("argument is not a valid address")
			logger.Warning("contining with default address of " + serverAddrDefault)
			serverAddr = serverAddrDefault
		}
		return serverAddr, nil
	}
	logger.Error("argument is not a valid address")
	logger.Warning("contining with default address of " + serverAddrDefault)
	serverAddr = serverAddrDefault
	return serverAddr, nil
}

func checkError(err error) {
	if err != nil {
		if logger != nil {
			logger.Fatal(err.Error())
		} else {
			fmt.Println(err.Error())
		}
		os.Exit(1)
	}
}

func printCommandLineUsageAndExit() {
	fmt.Println("USAGE: go run server.go PORT [LOCAL?]")
	os.Exit(1)
}
