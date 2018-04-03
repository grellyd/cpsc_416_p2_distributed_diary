package consensuslib

import (
	"consensuslib/paxosnode"
	"filelogger/singletonlogger"
	"fmt"
	"net"
	"net/rpc"
	"time"
)

// PaxosNodeRPCWrapper is the rpc wrapper around the paxos node
type PaxosNodeRPCWrapper = paxosnode.PaxosNodeRPCWrapper

// Client in the consensuslib
type Client struct {
	localAddr     string
	heartbeatRate time.Duration

	listener        net.Listener
	serverRPCClient *rpc.Client

	paxosNode           *paxosnode.PaxosNode
	paxosNodeRPCWrapper *PaxosNodeRPCWrapper
	neighbors           []string
}

// NewClient creates a new Client, ready to connect
func NewClient(clientAddr string, heartbeatRate time.Duration) (client *Client, err error) {
	client = &Client{
		heartbeatRate: heartbeatRate,
	}

	addr, err := net.ResolveTCPAddr("tcp", clientAddr)
	if err != nil {
		return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: unable to resolve client addr: %s", err)
	}

	client.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Unable to listen to IP address '%s': %s", addr, err)
	}
	client.localAddr = client.listener.Addr().String()
	singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#NewClient: Listening on IP address%v", client.localAddr))

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

// Connect the client to the server
func (c *Client) Connect(serverAddr string) (err error) {
	c.serverRPCClient, err = rpc.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to connect to server: %s", err)
	}
	singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#Connect: Registering to server at: %s\n", serverAddr))
	err = c.serverRPCClient.Call("Server.Register", c.localAddr, &c.neighbors)
	if err != nil {
		return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to register with server: %s", err)
	}
	go c.SendHeartbeats()

	if len(c.neighbors) > 0 {
		singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#Connect: Neighbors: %v\n", c.neighbors))
		err = c.paxosNode.SendNeighbours(c.neighbors)
		if err != nil {
			return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to connect to neighbors: %s", err)
		}
		singletonlogger.Debug("[LIB/CLIENT]#Connect: Learning the latest value from neighbours")
		err = c.paxosNode.LearnLatestValueFromNeighbours()
		log := c.paxosNode.Learner.Log
		if len(log) != 0 {
			rn := (log[len(log)-1].RoundNum) + 1
			c.paxosNode.SetRoundNum(rn)
		}
		//c.paxosNode.SetRoundNum(len(c.paxosNode.Learner.Log))


		if err != nil {
			return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to learn latest value while reading: %s", err)
		}
	}
	return nil
}

// TODO: change to display nicely
func (c *Client) Read() (value string, err error) {
	log, err := c.paxosNode.GetLog()
	if err != nil {
		return "", fmt.Errorf("[LIB/CLIENT]#Read: Error while getting the log: %s", err)
	}
	singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#Read: Log = '%v'\n", log))
	for _, m := range log {
		value += m.Value + "\n"
	}
	return value, nil
}

// TODO: Check for error
func (c *Client) Write(value string) (err error) {
	_, err = c.paxosNode.WriteToPaxosNode(value)
	return err
}

// IsAlive checks if the server is alive
func (c *Client) IsAlive() (alive bool, err error) {
	// alive is default false
	err = c.serverRPCClient.Call("Server.CheckAlive", c.localAddr, &alive)
	return alive, err
}

// TODO: use error log and continue

// SendHeartbeats to the server
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
