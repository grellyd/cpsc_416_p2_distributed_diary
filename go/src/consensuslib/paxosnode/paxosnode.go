package paxosnode

type PaxosNode struct {
	Addr			 string // IP:port, identifier
	Proposer   ProposerRole
	Acceptor   AcceptorRole
	Learner    LearnerRole
	Neighbours map[string]*rpc.client
}

type ProposerRole struct {
	
}