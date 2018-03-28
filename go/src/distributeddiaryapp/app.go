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
	"filelogger"
	"fmt"
	"os"
	"regexp"
	"time"
	"strconv"
)

var logger *filelogger.Logger
var validAddr = regexp.MustCompile("[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:[0-9]{1,5}")

const (
	serverAddrDefault = "127.0.0.1:12345"
	localAddrDefault  = "127.0.0.1:0"
)

func main() {
	var err error
	serverAddr, localAddr, err := parseArgs(os.Args)
	checkError(err)
	logger, err = filelogger.NewFileLogger("app", filelogger.NORMAL)
	checkError(err)
	logger.Debug("starting application")
	client, err := consensuslib.NewClient(localAddr, 1*time.Millisecond, logger)
	checkError(err)
	logger.Debug("created client")
	err = client.Connect(serverAddr)
	checkError(err)

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

	logger.Debug("serving")
	serveCli(client)
}

func setup() *consensuslib.Client {
	serverAddr := ""
	localPort := 0
	isLocal := false

	// Validate arguments
	if len(os.Args[1:]) == 2 {
		// Local arg not included; we're using a public IP
		serverAddr = os.Args[1]
		intPort, err := strconv.Atoi(os.Args[2])
		if err != nil {
			printCommandLineUsageAndExit()
		}
		localPort = intPort
	} else if len(os.Args[1:]) == 3 {
		// Local arg included; we're running this on 127.0.0.1
		serverAddr = os.Args[1]
		intPort, err := strconv.Atoi(os.Args[2])
		if err != nil {
			printCommandLineUsageAndExit()
		}
		localPort = intPort
		isLocal = true
	} else {
		printCommandLineUsageAndExit()
	}

	client, err := consensuslib.NewClient(localPort, isLocal, 1*time.Millisecond)
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
// 			fmt.Printf("[DD APP] Alive: %v\n", isAlive)
// 		case cli.EXIT:
// 			fmt.Println("[DD APP] Closing the Chamber of Secrets...")
// 			fmt.Println("[DD APP] Goodbye!")
// 			os.Exit(0)
// 		case cli.READ:
// 			value, err := client.Read()
// 			checkError(err)
// 			fmt.Printf("[DD APP] Reading: '%s'\n", value)
			logger.Info(fmt.Sprintf("Alive: %v", isAlive))
		case cli.EXIT:
			Exit()
		case cli.READ:
			value, err := client.Read()
			checkError(err)
			logger.Info(fmt.Sprintf("Reading: '%s'", value))
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
	logger.Info("Closing the Chamber of Secrets...")
	logger.Info("Goodbye!")
	logger.Exit()
	os.Exit(0)
}

func parseArgs(args []string) (serverAddr string, localAddr string, err error) {
	serverAddr = args[1]
	localAddr = args[2]
	if !validAddr.MatchString(serverAddr) || !validAddr.MatchString(localAddr) {
		logger.Error("arguments are not valid addresses")
		logger.Warning("contining with default addresses")
		serverAddr = serverAddrDefault
		localAddr = localAddrDefault
	}
	return serverAddr, localAddr, nil
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
	fmt.Println("USAGE: go run app.go SERVERIP:PORT LOCALPORT [isLocal?]")
	os.Exit(1)
}
