package main

import (
	"fmt"
	"os"
	"net"
	"net/rpc"
	"time"
)

var locAddr string
var serverConnector *rpc.Client
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
