package acceptor

import (
	"consensuslib"
)

type Message = consensuslib.Message

type AcceptorRole struct {
	LastPromised Message
	LastAccepted Message
}

func NewAcceptor() AcceptorRole {
	acc := AcceptorRole{nil, nil}
	return acc
}

type AcceptorInterface interface {

	// REQUIRES: a message with the empty/nil/'' string as a value;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	processPrepare (msg Message) Message

	// REQUIRES: a message with a value submitted at proposer;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	processAccept (msg Message) Message

}




