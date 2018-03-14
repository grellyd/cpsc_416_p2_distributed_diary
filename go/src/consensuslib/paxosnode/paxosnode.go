package paxosnode

import (
	"paxosnode/proposer"
)
type PaxosNode struct {
	Addr			 string // IP:port, identifier
	Proposer   ProposerRole
	Acceptor   AcceptorRole
	Learner    LearnerRole
	Neighbours map[string]*rpc.client
}

