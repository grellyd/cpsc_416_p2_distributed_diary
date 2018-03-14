package paxosnodeinterface

import (
	. "consensuslib"
)

// Handles the entire process of proposing a value and trying to achieve consensus
//TODO[sharon]: update parameters as needed.
func (pn *PaxosNode) WriteToPaxosNode(value string) (success bool, err error) {
	prepReq := pn.Proposer.CreatePrepareRequest()
	numAccepted, err := DisseminateRequest(prepReq) //TODO[sharon]: do error checking


	if !pn.IsMajority(numAccepted) {
		// TODO[sharon]: Handle not-majority. Quit or retry?
	}

	accReq := pn.Proposer.CreateAcceptRequest(value)
	numAccepted, err = DisseminateRequest(accReq)

	if !pn.IsMajority(numAccepted) {
		// TODO[sharon]: Handle not-majority. Quit or retry?
	}

	accReq.Type = CONSENSUS
	_, err = DisseminateRequest(accReq)

	return success, err
}

	// Sets up bidirectional RPC with all neighbours, given to the paxosnode by the client
func BecomeNeighbours(ips []string) (connectedNbrs []string, err error) {
	return connectedNbrs, err
}

// Sends a prepare request to all neighbours on behalf of the Paxosnode's proposer
// TODO[sharon]: Check parameters that get passed in

// Sends the value that consensus has been reached on to the entire network.
// Must be called after ProposeValue has returned successfully
//TODO[sharon]: Figure out best name for number field and add as param. Might be RPC
func	DisseminateRequest(prepReq Message) (numAccepted int, err error) {
	if prepReq.Type == ACCEPT {

	}
	return numAccepted, err
}

// Locally accepts the accept request sent by a PN in the system.
// TODO[sharon]: Figure out parameters. Might be RPC
func	AcceptAcceptRequest() (err error) {
	return err
}


	
func (pn *PaxosNode) IsMajority(n int) bool {
	if n > len(pn.Neighbours) {
		return true
	}
	return false
}