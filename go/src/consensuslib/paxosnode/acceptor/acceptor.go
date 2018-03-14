package acceptor

import (
	"proj2_c6y8_f1l0b_l0j8_l5w8_n5w8/go/src/consensuslib"
)


type AcceptorRole struct {
	LastPromised consensuslib.Message
	LastAccepted consensuslib.Message
}

func NewAcceptor() AcceptorRole {
	acc := AcceptorRole{
		nil,
		nil,
		}
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

func (acceptor *AcceptorRole) processPrepare (msg consensuslib.Message) consensuslib.Message {
	// no any value had been proposed or n'>n
	// then n' == n and ID' == ID (basically same proposer distributed proposal twice)
	if &acceptor.LastPromised == nil || msg.ID > acceptor.LastPromised.ID{
		acceptor.LastPromised = msg
	} else if acceptor.LastPromised.ID == msg.ID && acceptor.LastPromised.FromProposerID == msg.FromProposerID {
		acceptor.LastPromised = msg
	}
	return acceptor.LastPromised
}

func (acceptor *AcceptorRole) processAccept (msg consensuslib.Message) consensuslib.Message {
	if &acceptor.LastAccepted == nil {
		if (msg.ID == acceptor.LastPromised.ID &&
			msg.FromProposerID == acceptor.LastPromised.FromProposerID) {
			acceptor.LastAccepted = msg
		}
	}

	return acceptor.LastAccepted

}



