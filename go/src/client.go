package main

import (
	"consensuslib"
	"consensuslib/paxosnode/paxosnodeinterface"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"
)

type PaxosNodeInstance int

type Message = consensuslib.Message

var locAddr string
var serverConnector *rpc.Client
var paxnode paxosnodeinterface.PaxosNode

func main() {
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
	neighbours := make([]string, 0)
	_ = serverConnector.Call("Nserver.Register", locAddr, &neighbours)
	fmt.Println("Neighbours ", neighbours)

	go doEvery(1*time.Millisecond, SendHeartbeat)

	// initializing a new PN
	paxnode, err := paxosnodeinterface.MountPaxosNode(locAddr)
	if err != nil {
		fmt.Println("Couldn't create a PN")
		return
	}
	pni := new(PaxosNodeInstance)
	rpc.Register(pni)
	// connect PN to the neighboursbours
	if len(neighbours) != 0 {
		err = paxnode.BecomeNeighbours(neighbours)
		if err != nil {
			fmt.Println("Cannot connect to any neighboursbours, ping Server")
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
	time.Sleep(15 * time.Second)
}

func doEvery(d time.Duration, f func(time.Time) error) error {
	for x := range time.Tick(d) {
		f(x)
	}
	return nil
}

func SendHeartbeat(t time.Time) (err error) {
	var ignored bool
	err = serverConnector.Call("Nserver.HeartBeat", locAddr, &ignored)
	if err != nil {
		return err
	}
	return nil
}

// RPCs for paxosnodes start here
func (paxnodei *PaxosNodeInstance) ProcessPrepareRequest(m Message, r *Message) (err error) {
	*r = paxnode.Acceptor.ProcessPrepare(m)
	return nil
}

func (paxnodei *PaxosNodeInstance) ProcessAcceptRequest(m Message, r *Message) (err error) {
	*r = paxnode.Acceptor.ProcessAccept(m)
	return nil
}

func (paxnodei *PaxosNodeInstance) ProcessLearnRequest(m Message, r *Message) (err error) {
	// TODO: after Larissa's implementation put something like:
	// TODO: paxnode.Learner.Learn(m)
	*r = paxnode.Acceptor.ProcessAccept(m)
	if m.Equals(r) {
		// TODO: add func to call RPC "NotifyAboutAccepted"
	}
	return nil
}

// RPC call which is called by node that tries to connect
func (paxnodei *PaxosNodeInstance) ConnectRemoteNeighbour(addr string, r *bool) (err error) {
	err = paxnode.AcceptNeighbourConnection(addr, r)
	return err
}

// RPC call from other node's Acceptor about value it accepted
func (paxnodei *PaxosNodeInstance) NotifyAboutAccepted(m *Message, r *bool) (err error) {
	paxnode.CountForNumAlreadyAccepted(m)
	return err
}