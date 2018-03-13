package proposer

import "./proj2_c6y8_f1l0b_l0j8_l5w8_n5w8/go/src/consensuslib"

type ProposerRole struct {
	proposerID string
	messageID uint64
	CurrentPrepareRequest Message
	CurrentAcceptRequest Message
}


type ProposerInterface interface {
	// The proposer chooses a new prepare request ID and creates a prepare request
	// to return to the PN
	createPrepareRequest() Message
	// This creates an accept request with the current prepare request ID and a candidate value for consensus
	// to return to the PN. The value passed in is either an arbitrary value of the application's choosing, or is
	// the value corresponding to the highest prepare request ID contained in the permission granted messages from other
	// acceptors
	createAcceptRequest(value string) Message

	
}

func (proposer *ProposerRole) createPrepareRequest() Message {
	return Message{}
}

func (proposer *ProposerRole) createAcceptRequest(value string) Message {
	return Message{}
}

// The constructor for a new ProposerRole object instance. A PN should only interact with just one
// ProposerRole instance at a time
func newProposer(proposerID string) ProposerRole {
	return ProposerRole{}
}