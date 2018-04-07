package paxosnodeinterface

import (
	"consensuslib/message"
)

type Message = message.Message

/**
* Methods to be implemented by PaxosNode.
* This is the interface that the rest of the library uses to talk to the Paxos Network.
*
**/
type PaxosNodeInterface interface {

	// Gets the entire log on the Paxos Network
	GetLog() (log []Message, err error)

	// Handles the entire process of proposing a value and trying to achieve consensus.
	// ttl represents the # of times it will retry a write before it goes to sleep.
	// It will either try forever or fail.
	WriteToPaxosNode(value, msgHash string, ttl int) (success bool, err error)

	// Sets up bidirectional RPC with all neighbours
	// Can return the following errors:
	// - NeighbourConnectionError when establishing RPC connection with a neighbour fails
	BecomeNeighbours(ips []string) (err error)

	// Retrieves all the neighbours' logs and chooses the right candidate
	LearnLatestValueFromNeighbours() (err error)

	// Exit the Paxos Network
	UnmountPaxosNode() (err error)
}
