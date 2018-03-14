package paxosnode

import (
	//"consensuslib/paxosnode/acceptor"
	//"consensuslib/paxosnode/learner"
	//"consensuslib/paxosnode/proposer"
)

	// Handles the entire process of proposing a value and trying to achieve consensus
	//TODO[sharon]: update parameters as needed. Might be RPC
	func (pn *PaxosNode) ProposeValue(value string) (success bool, err error) {

	}

		// Sets up bidirectional RPC with all neighbours, given to the paxosnode by the client
		BecomeNeighbours(ips []string) (connectedNbrs []string, err error)

		// Sends the value that consensus has been reached on to the entire network.
		// Must be called after ProposeValue has returned successfully
		//TODO[sharon]: Figure out best name for number field and add as param. Might be RPC
		DisseminateAcceptedValue(value string) (success bool, err error)
	
		// Locally accepts the accept request sent by a PN in the system.
		// TODO[sharon]: Figure out parameters. Might be RPC
		AcceptAcceptRequest() (err error)
	
		// Sends a prepare request to all neighbours on behalf of the Paxosnode's proposer
		// TODO[sharon]: Check parameters that get passed in
		SendPrepareRequest(value string) (err error)
	
		// Exits the Paxosnode network.
		LeaveNetwork()