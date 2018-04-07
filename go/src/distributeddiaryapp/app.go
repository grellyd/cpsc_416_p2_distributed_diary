// Entrypoint for the Distributed Diary Application
// This file can be run with 'go run distributeddiaryapp/app.go'
// Or do `cd distributeddiaryapp && go build && ./distributeddiaryapp`
// Or do `go install` then `distributeddiaryapp` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// USAGE: go run app.go SERVERIP:PORT LOCALPORT [isLocal?]
// Go Run Example (Dev): `go run distributeddiaryapp/app.go 127.0.0.1:12345 8080 --local` -- To run on 127.0.0.1:8080
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
	"os"
	"paxostracker"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var validArgs = regexp.MustCompile("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:[0-9]{1,5} [0-9]{1,5}( " + localFlag + ")*( " + debugFlag + ")*")
var breaked bool
var written bool
var breakState, killState string

const (
	debugFlag = "--debug"
	localFlag = "--local"
	usage     = `==================================================
The Chamber of Secrets: A Distributed Diary App
==================================================
Usage: go run app.go serverAddress PORT [options]

Server address must be of the form 255.255.255.255:12345

Valid options:

--local : run on local machine at 127.0.0.1 with the specified port
--debug : run with debugging turned on for verbose logging
`
)

func main() {
	// Parse command line arguments
	serverAddr, localAddr, outboundAddr, logstate, err := parseArgs(os.Args[1:])
	checkError(err)

	// Create our logger
	err = singletonlogger.NewSingletonLogger("app", logstate)
	checkError(err)
	singletonlogger.Debug("[LIB/APP] starting application at " + localAddr + " with outbound address " + outboundAddr)

	// Create a new ConsensusLib client
	client, err := consensuslib.NewClient(localAddr, outboundAddr, 1*time.Millisecond)
	checkError(err)
	singletonlogger.Debug("[LIB/APP] created client at " + localAddr)

	// Connect to the ConsensusLib server at serverAddr
	err = client.Connect(serverAddr)
	checkError(err)
	singletonlogger.Debug("[LIB/APP] connected to server at " + serverAddr)
	singletonlogger.Debug("[LIB/APP] serving cli")

	// Serve the CLI interface to the Distributed Diary app
	serveCli(client)
}

func serveCli(client *consensuslib.Client) {
	for {
		command := cli.Run()
		singletonlogger.Debug(fmt.Sprintf("[app] received command %v", command))
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
			if breaked && !written {
				written = true
			} else if breaked && written {
				singletonlogger.Info("This client is at a breakpoint. Please 'continue' before writing again.")
				break
			}
			value := ""
			for i, s := range *command.Data {
				// add spaces
				if i != len(*command.Data)-1 {
					value += s + " "
				} else {
					value += s
				}
			}
			go client.Write(value)
		case cli.BREAK:
			if breaked && !written {
				singletonlogger.Info("This client is ready to hit a breakpoint. Please 'continue' before pausing again.")
				break
			} else if breaked && written {
				singletonlogger.Info("This client is at a breakpoint. Please 'continue' before pausing again.")
				break
			}
			breaked = true
			data := *command.Data
			breakState = data[0]
			singletonlogger.Info("Breaking before next " + breakState)
			switch breakState {
			case cli.Prepare:
				go paxostracker.BreakNextPrepare()
			case cli.Propose:
				go paxostracker.BreakNextPropose()
			case cli.Learn:
				go paxostracker.BreakNextLearn()
			case cli.Idle:
				go paxostracker.BreakNextIdle()
			case cli.Custom:
				go paxostracker.BreakNextCustom()
			default:
				singletonlogger.Error(fmt.Sprintf("Couldn't identify '%s'", breakState))
				breaked = false
			}
		case cli.KILL:
			data := *command.Data
			killState = data[0]
			singletonlogger.Info("Killing before next " + killState)
			switch killState {
			case cli.Prepare:
				go paxostracker.KillNextPrepare()
			case cli.Propose:
				go paxostracker.KillNextPropose()
			case cli.Learn:
				go paxostracker.KillNextLearn()
			case cli.Idle:
				go paxostracker.KillNextIdle()
			case cli.Custom:
				go paxostracker.KillNextCustom()
			default:
				singletonlogger.Error(fmt.Sprintf("Couldn't identify '%s'", killState))
			}
		case cli.CONTINUE:
			breaked = false
			written = false
			singletonlogger.Info("Continuing...")
			go paxostracker.Continue()
		case cli.ROUNDS:
			singletonlogger.Info(paxostracker.AsTable())
		case cli.STEP:
			if !breaked {
				singletonlogger.Info("Unable to step: Not at a breakpoint!")
				break
			}
			switch breakState {
			case cli.Prepare:
				singletonlogger.Info("Breaking before next Propose")
				breakState = cli.Propose
				go paxostracker.BreakNextPropose()
				go paxostracker.Continue()
			case cli.Propose:
				singletonlogger.Info("Breaking before next Learn")
				breakState = cli.Learn
				go paxostracker.BreakNextLearn()
				go paxostracker.Continue()
			case cli.Learn:
				singletonlogger.Info("Breaking before next Idle")
				breakState = cli.Idle
				go paxostracker.BreakNextIdle()
				go paxostracker.Continue()
			case cli.Idle:
				singletonlogger.Info("Cannot step beyond Idle. Please 'continue'")
			default:
				singletonlogger.Error(fmt.Sprintf("Couldn't identify '%s'", breakState))
			}
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

func parseArgs(args []string) (serverAddr string, localAddr string, outboundAddr string, logstate state.State, err error) {
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
			serverAddr = args[i]
		case 1:
			port, err = strconv.Atoi(args[i])
			if err != nil {
				return serverAddr, localAddr, outboundAddr, logstate, fmt.Errorf("error while converting port: %s", err)
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
		localAddr = "127.0.0.1" + addrEnd
		outboundAddr = "127.0.0.1" + addrEnd
	} else {
		outboundIP, err := networking.GetOutboundIP()
		if err != nil {
			return serverAddr, localAddr, outboundAddr, logstate, fmt.Errorf("error while fetching ip: %s", err)
		}
		outboundAddr = outboundIP + addrEnd
		localAddr = addrEnd

	}
	return serverAddr, localAddr, outboundAddr, logstate, nil
}

func checkError(err error) {
	if err != nil {
		singletonlogger.Fatal(err.Error())
		os.Exit(1)
	}
}
