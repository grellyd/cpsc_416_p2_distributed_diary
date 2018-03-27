// Entrypoint for the Distributed Diary Server
// This file can be run with 'go run distributeddiaryserver/server.go'
// Or do `cd distributeddiaryserver && go build && ./distributeddiaryserver`
// Or do `go install` then `distributeddiaryserver` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

// Go Run Example: `go run distributeddiaryserver/server.go 127.0.0.1:12345`
// Installed Run example: `distributeddiaryserver 127.0.0.1:12345`

package main

import (
	"consensuslib"
	"fmt"
	"os"
	"networking"
)

func main() {
	fmt.Println("testing server")
	addr := "127.0.0.1:12345"
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
