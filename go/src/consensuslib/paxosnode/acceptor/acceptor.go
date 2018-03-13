package acceptor

import (
	"proj2/consensuslib/paxosnode"
)

type Acceptor struct {
	LastPromised Message
	LastAccepted Message
}

func NewAcceptor() Acceptor {
	acc := Acceptor{nil, nil}
	return acc
}


