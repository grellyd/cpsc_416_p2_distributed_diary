package proposer

import (
	"consensuslib"
)

type Message = consensuslib.Message

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
	// This is used by the PN to inform its proposer of the highest message ID value it has seen
	// so far from other PNs. All future prepare requests must have a messageID greater than
	// the messageID passed in
	updateMessageID(messageID uint64)

	
}

func (proposer *ProposerRole) createPrepareRequest() Message {
	// Increment the messageID (n value) every time a new prepare request is made
	proposer.messageID++
	prepareRequest := Message{
		ID: proposer.messageID,
		MsgType: "prepare",
		Value: "",
		FromProposerID: proposer.proposerID,
	}
	return prepareRequest
}

func (proposer *ProposerRole) createAcceptRequest(value string) Message {
	acceptRequest := Message{
		ID: proposer.messageID,
		MsgType: "accept",
		Value: value,
		FromProposerID: proposer.proposerID,
	}
	return acceptRequest
}

func (proposer *ProposerRole) updateMessageID(messageID uint64) {
	proposer.messageID = messageID
}

// The constructor for a new ProposerRole object instance. A PN should only interact with just one
// ProposerRole instance at a time
func newProposer(proposerID string) *ProposerRole {
	proposer := &ProposerRole{
		proposerID: proposerID,
		messageID: 0,
		CurrentPrepareRequest: nil,
		CurrentAcceptRequest: nil,
	}
	return proposer
}