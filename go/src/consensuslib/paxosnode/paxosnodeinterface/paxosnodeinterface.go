package paxosnodeinterface

type PaxosNodeInterface interface {
	// Gets the entire log on the PN
	ReadFromPaxosNode() (err error)

	// Tries to get the value given written into the log
	WriteToPaxosNode(value string) (err error)

	// Passes the list of neighbour addresses to the PN
	// Can return the following errors:
	// - NeighbourConnectionError when establishing RPC connection with a neighbour fails
	SendNeighbours(ips []string) (err error)

	LearnLatestValueFromNeighbours() (err error)

	// Exit the PN
	UnmountPaxosNode() (err error)
}
