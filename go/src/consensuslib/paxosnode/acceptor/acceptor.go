package acceptor

import (
	"consensuslib"
)


type AcceptorRole struct {
	LastPromised consensuslib.Message
	LastAccepted consensuslib.Message
}

func NewAcceptor() AcceptorRole {
	acc := AcceptorRole{nil, nil}
	return acc
}

type AcceptorInterface interface {

	// REQUIRES: a message with the empty/nil/'' string as a value;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	processPrepare (msg consensuslib.Message) consensuslib.Message

	// REQUIRES: a message with a value submitted at proposer;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	processAccept (msg consensuslib.Message) consensuslib.Message

}




