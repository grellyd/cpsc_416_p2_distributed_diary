package paxosnodeinterface

import (
	"consensuslib/paxosnode/acceptor"
	"consensuslib/paxosnode/learner"
	"consensuslib/paxosnode/proposer"
	"net/rpc"
)

type ProposerRole = proposer.ProposerRole
type AcceptorRole = acceptor.AcceptorRole
type LearnerRole = learner.LearnerRole

type PaxosNode struct {
	Addr       string // IP:port, identifier
	Proposer   ProposerRole
	Acceptor   AcceptorRole
	Learner    LearnerRole
	NbrAddrs   []string
	Neighbours map[string]*rpc.Client
}

type PaxosNodeInterface interface {
	// Gets the entire log on the PN
	ReadFromPaxosNode() (err error)

	// TODO[sharon]: Might not include this function. Reads from a specific index.
	ReadAtFromPaxosNode() (err error)

	// Tries to get the value given written into the log
	WriteToPaxosNode(value string) (err error)

	// Passes the list of neighbour addresses to the PN
	SendNeighbours([]string) (err error)

	// Exit the PN
	UnmountPaxosNode() (err error)
}

// A client will call this to mount to the PN
// TODO[sharon]TODO[all]: Implement
func MountPaxosNode(pnAddr string) (err error) {
	return err
}

