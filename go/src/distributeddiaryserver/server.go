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
	"fmt"
	//"consensuslib"
	"filelogger/singletonlogger"
	"filelogger/state"
	"os"
	"strings"
	"strconv"
	"regexp"
)

const (
	localFlag = "--local"
	debugFlag = "--debug"
	usage = `==================================================
The Chamber of Secrets: A Distributed Diary Server
==================================================
Usage: go run server.go PORT [options]

Valid options:

--local : run on local machine at 127.0.0.1 with the specified port
--debug : run with debuggging turned on for verbose logging
`
)

var validArgs = regexp.MustCompile("[0-9]{1,5}( " + localFlag + ")*( " + debugFlag +")*")

func main() {
	port, logstate, isLocal, err := parseArgs(os.Args[1:])
	checkError(err)
	err = singletonlogger.NewSingletonLogger("server", logstate)
	checkError(err)
	singletonlogger.Debug("Logger created")
	addr := setAddr(port, isLocal)
	singletonlogger.Debug("Chosen Addr: " + addr)
	singletonlogger.Debug("Creating consensuslib server for " + addr)
	//server, err := consensuslib.NewServer(addr, logger)
	checkError(err)
	singletonlogger.Info("Serving at " + addr)
	//err = server.Serve()
	checkError(err)
}

func parseArgs(args []string) (port int, logstate state.State, isLocal bool, err error) {
	if !validArgs.MatchString(strings.Join(args, " ")) {
		fmt.Println(usage)
		os.Exit(1)
	}
	for i, arg := range(args) {
		// positional args
		switch i {
		case 0: 
		port, err = strconv.Atoi(args[0])
		if err != nil {
			return port, logstate, isLocal, fmt.Errorf("error while converting port: %s", err)
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
	return port, logstate, isLocal, nil
}

func setAddr(port int, isLocal bool) (addr string) {
	addrEnd := fmt.Sprintf(":%d", port)
	if isLocal {
		addr = "127.0.0.1:" + addrEnd
	} else {
		addr = addrEnd
	}
	return addr
}

func checkError(err error) {
	if err != nil {
		singletonlogger.Fatal(err.Error())
		os.Exit(1)
	}
}
