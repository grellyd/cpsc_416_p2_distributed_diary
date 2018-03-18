package acceptor

import (
	. "consensuslib"
)

type AcceptorRole struct {
	LastPromised Message
	LastAccepted Message
}

func NewAcceptor() AcceptorRole {
	acc := AcceptorRole{
		Message{},
		Message{},
	}
	return acc
}

type AcceptorInterface interface {

	// REQUIRES: a message with the empty/nil/'' string as a value;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	ProcessPrepare(msg Message) Message

	// REQUIRES: a message with a value submitted at proposer;
	// EFFECTS: responds with the latest promised/accepted message or with the nil if none
	ProcessAccept(msg Message) Message
}

func (acceptor *AcceptorRole) ProcessPrepare(msg Message) Message {
	// no any value had been proposed or n'>n
	// then n' == n and ID' == ID (basically same proposer distributed proposal twice)
	if &acceptor.LastPromised == nil || msg.ID > acceptor.LastPromised.ID {
		acceptor.LastPromised = msg
	} else if acceptor.LastPromised.ID == msg.ID && acceptor.LastPromised.FromProposerID == msg.FromProposerID {
		acceptor.LastPromised = msg
	}
	return acceptor.LastPromised
}

func (acceptor *AcceptorRole) ProcessAccept(msg Message) Message {
	if &acceptor.LastAccepted == nil {
		if msg.ID == acceptor.LastPromised.ID &&
			msg.FromProposerID == acceptor.LastPromised.FromProposerID {
			acceptor.LastAccepted = msg
		} else if msg.ID > acceptor.LastPromised.ID {
			//acceptor.LastPromised = msg
			acceptor.LastAccepted = msg
		}
	} else {
		if msg.ID == acceptor.LastPromised.ID &&
			acceptor.LastPromised.FromProposerID == msg.FromProposerID {
			acceptor.LastAccepted = msg
		} else if msg.ID > acceptor.LastPromised.ID || msg.ID > acceptor.LastAccepted.ID {
			acceptor.LastAccepted = msg
		}
	}

	return acceptor.LastAccepted

}
