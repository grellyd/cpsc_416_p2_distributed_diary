package paxosnode

import (
	"consensuslib/paxosnode/proposer";
	"consensuslib/paxosnode/acceptor";
	"consensuslib/paxosnode/learner"
)

type ProposerRole = proposer.ProposerRole
type AcceptorRole = acceptor.AcceptorRole
type LearnerRole = learner.LearnerRole

type PaxosNode struct {
	Addr			 string // IP:port, identifier
	Proposer   ProposerRole
	Acceptor   AcceptorRole
	Learner    LearnerRole
	Neighbours map[string]*rpc.client
}

type PaxosNodeInterface interface {

}
