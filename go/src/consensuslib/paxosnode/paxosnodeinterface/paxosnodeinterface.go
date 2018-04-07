package paxosnodeinterface

/**
* Methods to be implemented by PaxosNode.
* This is the interface that the rest of the library uses to talk to the Paxos Network.
*
**/
type PaxosNodeInterface interface {

	// Gets the entire log on the Paxos Network
	ReadFromPaxosNode() (err error)

	// Tries to get the value given written into the log
	WriteToPaxosNode(value string) (err error)

	// Passes the list of neighbour addresses to the PN
	// Can return the following errors:
	// - NeighbourConnectionError when establishing RPC connection with a neighbour fails
	BecomeNeighbours(ips []string) (err error)

	// Retrieves all the neighbours' logs and chooses the right candidate
	LearnLatestValueFromNeighbours() (err error)

	// Exit the Paxos Network
	UnmountPaxosNode() (err error)
}
