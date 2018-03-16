package main

import (
	"fmt"
	"os"
	"net"
	"net/rpc"
	"time"
	"proj2_c6y8_f1l0b_l0j8_l5w8_n5w8/go/src/consensuslib/paxosnode/paxosnodeinterface"
)

var locAddr string
var serverConnector *rpc.Client
var paxnode paxosnodeinterface.PaxosNode

func main()  {
	fmt.Println("start")
	args := os.Args[1:]

	servAddr := args[0]
	localIP := "127.0.0.1:0"

	serverAddr, _ := net.ResolveTCPAddr("tcp", servAddr)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", localIP)
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	locAddr = listener.Addr().String()
	fmt.Println("Local addr ", locAddr)

	serverConnector, _ = rpc.Dial("tcp", serverAddr.String())
	neigh := make([]string,0)
	_ = serverConnector.Call("Nserver.Register", locAddr, &neigh)
	fmt.Println("Neighbours ", neigh)

	go doEvery(1 * time.Millisecond, SendHeartbeat)

	// initializing a new PN
	paxnode, err := paxosnodeinterface.MountPaxosNode(locAddr)
	if err != nil {
		fmt.Println("Couldn't create a PN")
		return
	}
	// connect PN to the neighbours
	if len(neigh) != 0 {
		err = paxnode.BecomeNeighbours(neigh)
		if err != nil {
			fmt.Println("Cannot connect to any neighbours, ping Server")
			// ping server here whether we're alive
			alive := false
			err = serverConnector.Call("Nserver.CheckAlive", locAddr, &alive)
			if err != nil {
				fmt.Println("Client disconnected from the net")
				return
			}
		}
	}

	// TODO: wait for the commands from the app

	fmt.Println("Sleeping now")
	time.Sleep(15*time.Second)
}

func doEvery(d time.Duration, f func(time.Time) error) error {
	for x := range time.Tick(d) {
		f(x)
	}
	return nil
}

func SendHeartbeat(t time.Time) (err error) {
	var ignored bool
	err = serverConnector.Call( "Nserver.HeartBeat", locAddr, &ignored)
	if err != nil {
		return err
	}
	return nil
}
