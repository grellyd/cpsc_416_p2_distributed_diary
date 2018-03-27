package consensuslib

import (
	"consensuslib/paxosnode"
	"consensuslib/util/networking"
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


// TODO: pass in logger
func NewClient(localPort int, heartbeatRate time.Duration) (client *Client, err error) {
	/****
	Creates a new client so that it is ready to connect to the server.

	Set `localPort` to a valid port # if the server is running on the same machine (e.g. both are running on 127.0.0.1).
	Otherwise, set it to < 0 in production, and the client will use the public outbound IP to register with the server.
	****/
	client = &Client{
		heartbeatRate: heartbeatRate,
	}

	addr := &net.TCPAddr{}
	if localPort >= 0 {
		localAddr := fmt.Sprintf("127.0.0.1:%d", localPort)
		addr, err = net.ResolveTCPAddr("tcp", localAddr)
		if err != nil {
			return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Unable to resolve local address '%s': %s", localAddr, err)
		}
	} else {
		publicAddr, err := networking.GetOutboundIP()
		if err != nil {
			return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Outbound IP couldn't be fetched")
		}
		addr, err = net.ResolveTCPAddr("tcp", publicAddr)
		if err != nil {
			return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Unable to resolve a public address: %s", err)
		}
	}

	client.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Unable to listen to IP address '%s': %s", addr, err)
	}
	client.localAddr = client.listener.Addr().String()
	fmt.Println("[LIB/CLIENT]#NewClient: Listening on IP address", client.localAddr)

	// create the paxosnode
	client.paxosNode, err = paxosnode.NewPaxosNode(client.localAddr)
	if err != nil {
		return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Unable to create a paxos node: %s", err)
	}

	// add the rpc wrapper
	client.paxosNodeRPCWrapper, err = paxosnode.NewPaxosNodeRPCWrapper(client.paxosNode)
	if err != nil {
		return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Unable to create RPC wrapper: %s", err)
	}
	rpc.Register(client.paxosNodeRPCWrapper)
	go rpc.Accept(client.listener)

	return client, nil
}

func (c *Client) Connect(serverAddr string) (err error) {
	c.serverRPCClient, err = rpc.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to connect to server: %s", err)
	}
	err = c.serverRPCClient.Call("Server.Register", c.localAddr, &c.neighbors)
	if err != nil {
		return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to register with server: %s", err)
	}
	go c.SendHeartbeats()

	if len(c.neighbors) > 0 {
		fmt.Printf("[LIB/CLIENT]#Connect: Neighbors: %v\n", c.neighbors)
		err = c.paxosNode.SendNeighbours(c.neighbors)
		if err != nil {
			return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to connect to neighbors: %s", err)
		}
		err = c.paxosNode.LearnLatestValueFromNeighbours()
		if err != nil {
			return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to learn latest value while reading: %s", err)
		}
	}
	return nil
}

// TODO
func (c *Client) Read() (value string, err error) {
	log, err := c.paxosNode.GetLog()
	if err != nil {
		return "", fmt.Errorf("[LIB/CLIENT]#Read: Error while getting the log: %s", err)
	}
	fmt.Printf("[LIB/CLIENT]#Read: Log = '%v'\n", log)
	for _, m := range log {
		value += m.Value
	}
	return value, nil
}

// TODO: Check for error
func (c *Client) Write(value string) (err error) {
	_, err = c.paxosNode.WriteToPaxosNode(value)
	return err
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
			return fmt.Errorf("[LIB/CLIENT]#SendHeartheats: Error while sending heartbeat: %s", err)
		}
	}
	return nil
}
