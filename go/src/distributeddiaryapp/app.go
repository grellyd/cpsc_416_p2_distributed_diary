// Entrypoint for the Distributed Diary Application
// This file can be run with 'go run distributeddiaryapp/app.go'
// Or do `cd distributeddiaryapp && go build && ./distributeddiaryapp`
// Or do `go install` then `distributeddiaryapp` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// USAGE: go run app.go SERVERIP:PORT LOCALPORT [isLocal?]
// Go Run Example (Dev): `go run distributeddiaryapp/app.go 127.0.0.1:12345 8080 LOCAL` -- To run on 127.0.0.1:8080
// Go Run Example (Prod): `go run distributeddiaryapp/app.go 127.0.0.1:12345 8080` -- To run on machine's outbound IP on port 8080
// Installed Run example: `distributeddiaryapp 127.0.0.1:12345 8080`

package main

import (
	"consensuslib"
	"distributeddiaryapp/cli"
	"distributeddiaryapp/networking"
	"filelogger/singletonlogger"
	"filelogger/state"
	"fmt"
	"time"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var validArgs = regexp.MustCompile("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:[0-9]{1,5} [0-9]{1,5}( " + localFlag + ")*( " + debugFlag +")*")

const (
	serverAddrDefault = "127.0.0.1:12345"
	localAddrDefault  = "127.0.0.1:0"
	debugFlag = "--debug"
	localFlag = "--local"
	usage = `==================================================
The Chamber of Secrets: A Distributed Diary App
==================================================
Usage: go run app.go serverAddress PORT [options]

Server address must be of the form 255.255.255.255:12345

Valid options:

--local : run on local machine at 127.0.0.1 with the specified port
--debug : run with debuggging turned on for verbose logging
`
)

func main() {
	serverAddr, localAddr, logstate, err := parseArgs(os.Args[1:])
	checkError(err)
	err = singletonlogger.NewSingletonLogger("app", logstate)
	checkError(err)
	singletonlogger.Debug("starting application at " + localAddr)
	client, err := consensuslib.NewClient(localAddr, 1*time.Millisecond)
	checkError(err)
	singletonlogger.Debug("created client at " + localAddr)
	err = client.Connect(serverAddr)
	checkError(err)
	singletonlogger.Debug("connected to server at " + serverAddr)
	singletonlogger.Debug("serving cli")
	serveCli(client)
}

func serveCli(client *consensuslib.Client) {
	for {
		command := cli.Run()
		switch command.Command {
		case cli.ALIVE:
			isAlive, err := client.IsAlive()
			checkError(err)
			singletonlogger.Info(fmt.Sprintf("Alive: %v", isAlive))
		case cli.EXIT:
			Exit()
		case cli.READ:
			value, err := client.Read()
			checkError(err)
			singletonlogger.Info(fmt.Sprintf("Reading: \n%s", value))
		case cli.WRITE:
			value := ""
			for i, s := range *command.Data {
				// add spaces
				if i != len(*command.Data)-1 {
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

// Exit nicely from the program
func Exit() {
	// TODO: Delete temp folder
	singletonlogger.Info("Closing the Chamber of Secrets...")
	singletonlogger.Info("Goodbye!")
	os.Exit(0)
}

func parseArgs(args []string) (serverAddr string, clientAddr string, logstate state.State, err error) {
	if !validArgs.MatchString(strings.Join(args, " ")) {
		fmt.Println(usage)
		os.Exit(1)
	}
	port := 0
	isLocal := false
	for i, arg := range(args) {
		// positional args
		switch i {
		case 0:
			serverAddr = args[i]
		case 1: 
		port, err = strconv.Atoi(args[i])
		if err != nil {
			return serverAddr, clientAddr, logstate, fmt.Errorf("error while converting port: %s", err)
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
		clientAddr = "127.0.0.1" + addrEnd
	} else {
		ip, err := networking.GetOutboundIP()
		if err != nil {
			return serverAddr, clientAddr, logstate, fmt.Errorf("error while fetching ip: %s", err)
		}
		clientAddr = ip + addrEnd
	}
	return serverAddr, clientAddr, logstate, nil
}

func checkError(err error) {
	if err != nil {
		singletonlogger.Fatal(err.Error())
		os.Exit(1)
	}
}
