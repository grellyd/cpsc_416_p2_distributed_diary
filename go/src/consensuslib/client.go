package consensuslib

import (
	"consensuslib/paxosnode"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

type PaxosNodeRPCWrapper = paxosnode.PaxosNodeRPCWrapper

type Client struct {
	localAddr     string
	heartbeatRate time.Duration

	listener        net.Listener
	serverRPCClient *rpc.Client

	paxosNode           *paxosnode.PaxosNode
	paxosNodeRPCWrapper *PaxosNodeRPCWrapper
	neighbors           []string

	errLog *log.Logger
	outLog *log.Logger
}

// Creates a new Client, ready to connect
// TODO: pass in logger
func NewClient(localAddr string, heartbeatRate time.Duration) (client *Client, err error) {
	client = &Client{
		heartbeatRate: heartbeatRate,
	}
	// in order to get out local ip and pick a port, we must assign a listener
	addr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve localaddr '%s': %s", localAddr, err)
	}
	client.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("unable to listen to localaddr '%s': %s", localAddr, err)
	}
	client.localAddr = client.listener.Addr().String()
	fmt.Println("Local addr ", client.localAddr)

	// create the paxosnode
	client.paxosNode, err = paxosnode.NewPaxosNode(client.localAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to create a paxos node: %s", err)
	}

	// add the rpc wrapper
	client.paxosNodeRPCWrapper, err = paxosnode.NewPaxosNodeRPCWrapper(client.paxosNode)
	if err != nil {
		return nil, fmt.Errorf("unable to create rpc wrapper: %s", err)
	}
	rpc.Register(client.paxosNodeRPCWrapper)
	go rpc.Accept(client.listener)

	return client, nil
}

func (c *Client) Connect(serverAddr string) (err error) {
	c.serverRPCClient, err = rpc.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("unable to connect to server: %s", err)
	}
	err = c.serverRPCClient.Call("Server.Register", c.localAddr, &c.neighbors)
	if err != nil {
		return fmt.Errorf("unable to register with server: %s", err)
	}
	go c.SendHeartbeats()

	if len(c.neighbors) > 0 {
		fmt.Printf("Neighbors: %v\n", c.neighbors)
		err = c.paxosNode.SendNeighbours(c.neighbors)
		if err != nil {
			return fmt.Errorf("unable to connect to neighbors: %s", err)
		}
	}
	return nil
}

// TODO
func (c *Client) Read() (value string, err error) {
	err = c.paxosNode.LearnLatestValueFromNeighbours()
	if err != nil {
		return "", fmt.Errorf("unable to learn latest value while reading: %s", err)
	}
	log, err := c.paxosNode.GetLog()
	if err != nil {
		return "", fmt.Errorf("error while getting the log: %s", err)
	}
	fmt.Printf("log: '%v'\n", log)
	for _, m := range log {
		value += m.Value
	}
	return value, nil
}

// TODO: Check for error
func (c *Client) Write(value string) (err error) {
	c.paxosNode.WriteToPaxosNode("hello")
	return nil
}

func (c *Client) IsAlive() (alive bool, err error) {
	// alive is default false
	err = c.serverRPCClient.Call("Server.CheckAlive", c.localAddr, &alive)
	return alive, err
}

// TODO: use error log and continue
func (c *Client) SendHeartbeats() (err error) {
	for _ = range time.Tick(c.heartbeatRate) {
		// TODO: Check ignored
		var ignored bool
		err = c.serverRPCClient.Call("Server.HeartBeat", c.localAddr, &ignored)
		if err != nil {
			return fmt.Errorf("error while sending heartbeat: %s", err)
		}
	}
	return nil
}
