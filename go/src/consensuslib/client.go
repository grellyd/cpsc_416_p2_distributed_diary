package consensuslib

import (
	"consensuslib/paxosnode"
	"filelogger/singletonlogger"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"paxostracker"
	"time"
)

// MSGHASHLEN Represents the length of message hash
const MSGHASHLEN = 4

// PaxosNodeRPCWrapper is the rpc wrapper around the paxos node
type PaxosNodeRPCWrapper = paxosnode.PaxosNodeRPCWrapper

// Client in the consensuslib
type Client struct {
	localAddr     string
	outboundAddr  string
	heartbeatRate time.Duration

	listener        net.Listener
	serverRPCClient *rpc.Client

	paxosNode           *paxosnode.PaxosNode
	paxosNodeRPCWrapper *PaxosNodeRPCWrapper
	neighbors           []string
}

// NewClient creates a new Client, ready to connect
func NewClient(localAddr string, outboundAddr string, heartbeatRate time.Duration) (client *Client, err error) {
	client = &Client{
		heartbeatRate: heartbeatRate,
	}

	addr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: unable to resolve client addr: %s", err)
	}

	client.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("[LIB/CLIENT]#NewClient: Unable to listen to IP address '%s': %s", addr, err)
	}
	client.localAddr = client.listener.Addr().String()
	client.outboundAddr = outboundAddr
	singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#NewClient: Listening on IP address %v", client.localAddr))
	singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#NewClient: Outbound IP address is %v", client.outboundAddr))

	// create the paxosnode
	client.paxosNode, err = paxosnode.NewPaxosNode(client.outboundAddr)
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

	paxostracker.NewPaxosTracker()
	return client, nil
}

// Connect the client to the server at serverAddr
func (c *Client) Connect(serverAddr string) (err error) {
	c.serverRPCClient, err = rpc.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to connect to server: %s", err)
	}

	// Register outboundAddr with the server so the server can 1) receive heartbeats, and 2) inform neighbours about us
	// The server will populate our neighbours field with our neighbours
	singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#Connect: Registering to server at: %s\n", serverAddr))
	err = c.serverRPCClient.Call("Server.Register", c.outboundAddr, &c.neighbors)
	if err != nil {
		return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to register with server: %s", err)
	}
	go c.SendHeartbeats()

	// For each neighbour received from the server, 1) set up a connection, and 2) Learn what log values they have.
	// Then, choose the longest log received from the neighbours. Lastly, set up the round number the network is
	// currently at.
	if len(c.neighbors) > 0 {
		singletonlogger.Debug(fmt.Sprintf("[LIB/CLIENT]#Connect: Neighbors: %v\n", c.neighbors))
		err = c.paxosNode.BecomeNeighbours(c.neighbors)
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

		if err != nil {
			return fmt.Errorf("[LIB/CLIENT]#Connect: Unable to learn latest value while reading: %s", err)
		}
	}
	return nil
}

// Read the node's version of the log
// It should be eventually consistent to the Paxos Network's agreed-upon version of the log.
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

// Write to the shared log
func (c *Client) Write(value string) (err error) {
	paxostracker.Prepare(c.listener.Addr().String())
	messageHash := generateMessageHash(MSGHASHLEN)
	_, err = c.paxosNode.WriteToPaxosNode(value, messageHash, paxosnode.TTL)
	return err
}

// IsAlive checks if the server is alive
func (c *Client) IsAlive() (alive bool, err error) {
	// alive is default false
	err = c.serverRPCClient.Call("Server.CheckAlive", c.outboundAddr, &alive)
	return alive, err
}

// SendHeartbeats to the server
func (c *Client) SendHeartbeats() (err error) {
	for _ = range time.Tick(c.heartbeatRate) {
		var ignored bool
		err = c.serverRPCClient.Call("Server.HeartBeat", c.outboundAddr, &ignored)
		if err != nil {
			return fmt.Errorf("[LIB/CLIENT]#SendHeartheats: Error while sending heartbeat: %s", err)
		}
	}
	return nil
}

func generateMessageHash(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
