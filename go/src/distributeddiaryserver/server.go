// Entrypoint for the Distributed Diary Server
// This file can be run with 'go run distributeddiaryserver/server.go'
// Or do `cd distributeddiaryserver && go build && ./distributeddiaryserver`
// Or do `go install` then `distributeddiaryserver` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// Go Run Example: `go run distributeddiaryserver/server.go 12345 --local` -- To run server on 127.0.0.1:12345
// Go Run Example: `go run distributeddiaryserver/server.go 12345` -- To run server on the outbound IP address, on port 12345
// Installed Run example: `distributeddiaryserver 12345`

package main

import (
	"consensuslib"
	"filelogger/singletonlogger"
	"filelogger/state"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	localFlag = "--local"
	debugFlag = "--debug"
	usage     = `==================================================
The Chamber of Secrets: A Distributed Diary Server
==================================================
Usage: go run server.go PORT [options]

Valid options:

--local : run on local machine at 127.0.0.1 with the specified port
--debug : run with debuggging turned on for verbose logging
`
)

var validArgs = regexp.MustCompile("[0-9]{1,5}( " + localFlag + ")*( " + debugFlag + ")*")

func main() {
	addr, logstate, err := parseArgs(os.Args[1:])
	checkError(err)
	err = singletonlogger.NewSingletonLogger("server", logstate)
	checkError(err)
	singletonlogger.Debug("Logger created")
	singletonlogger.Debug("Chosen Addr: " + addr)
	singletonlogger.Debug("Creating consensuslib server for " + addr)
	server, err := consensuslib.NewServer(addr)
	checkError(err)
	singletonlogger.Info("Serving at " + addr)
	err = server.Serve()
	checkError(err)
}

func parseArgs(args []string) (addr string, logstate state.State, err error) {
	if !validArgs.MatchString(strings.Join(args, " ")) {
		fmt.Println(usage)
		os.Exit(1)
	}
	port := 0
	isLocal := false
	for i, arg := range args {
		// positional args
		switch i {
		case 0:
			port, err = strconv.Atoi(args[i])
			if err != nil {
				return addr, logstate, fmt.Errorf("error while converting port: %s", err)
			}
		default:
			// option flags
			switch arg {
			case localFlag:
				isLocal = true
			case debugFlag:
				logstate = state.DEBUGGING
			}
		}
	}
	addrEnd := fmt.Sprintf(":%d", port)
	if isLocal {
		addr = "127.0.0.1" + addrEnd
	} else {
		addr = addrEnd
	}
	return addr, logstate, nil
}

func checkError(err error) {
	if err != nil {
		singletonlogger.Fatal(err.Error())
		os.Exit(1)
	}
}
