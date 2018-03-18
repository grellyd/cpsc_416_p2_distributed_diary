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

	// TODO[sharon]TODO[all]: Might not include this function. Reads from a specific index.
	ReadAtFromPaxosNode() (err error)

	// Tries to get the value given written into the log
	WriteToPaxosNode(value string) (err error)

	// Passes the list of neighbour addresses to the PN
	// Can return the following errors:
	// - NeighbourConnectionError when establishing RPC connection with a neighbour fails
	SendNeighbours(ips []string) (err error)

	// Exit the PN
	UnmountPaxosNode() (err error)
}

func (pn *PaxosNode) SendNeighbours(ips []string) (err error) {
	err = pn.BecomeNeighbours(ips)
	return err
}

func (pn *PaxosNode) UnmountPaxosNode() (err error) {
	// Close all RPC connections with neighbours during unmount
	for _, conn := range pn.Neighbours {
		conn.Close()
	}
	pn.NbrAddrs = nil

	return nil
}

// A client will call this to mount to create a Paxos Node that
// is linked to the client. The PN's Addr field is set as the pnAddr passed in
func MountPaxosNode(pnAddr string) (pn PaxosNode, err error) {
	proposer := proposer.NewProposer(pnAddr)
	acceptor := acceptor.NewAcceptor()
	learner := learner.NewLearner()
	pn = PaxosNode{
		Addr:     pnAddr,
		Proposer: proposer,
		Acceptor: acceptor,
		Learner:  learner,
	}
	return pn, err
}
