package paxosnode

import (
	"consensuslib/paxosnode/acceptor"
	"consensuslib/paxosnode/learner"
	"consensuslib/paxosnode/proposer"
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
	Neighbours map[string]*rpc.client
}

type PaxosNodeInterface interface {
	// Sets up bidirectional RPC with all neighbours, given to the paxosnode by the client
	BecomeNeighbours(ips []string) (connectedNbrs []string, err error)

	// Handles the entire process of proposing a value and trying to achieve consensus
	//TODO[]sharon: update parameters as needed.
	ProposeValue(value string) (success bool, err error)

	// Sends the value that consensus has been reached on to the entire network.
	// Must be called after ProposeValue has returned successfully
	//TODO[sharon]: Figure out best name for number field and add as param
	DisseminateAcceptedValue(value string) (success bool, err error)

	// Locally accepts the accept request sent by a PN in the system.
	// TODO[sharon]: Figure out parameters
	AcceptAcceptRequest() (err error)

	// Exits the Paxosnode network.
	LeaveNetwork()
}
