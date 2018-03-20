// Entrypoint for the Distributed Diary Server
// This file can be run with 'go run distributeddiaryserver/server.go'
// Or do `cd distributeddiaryserver && go build && ./distributeddiaryserver`
// Or do `go install` then `distributeddiaryserver` to run the binary
// The last is @grellyd preferred for ease, but requires you to add `go/bin` to your $PATH variable

package main

import (
	"fmt"
	"os"
	"consensuslib"
)

func main() {
	fmt.Println("testing server")
	addr := "127.0.0.1:12345"
	server, err := consensuslib.NewServer(addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	server.Serve()
}
