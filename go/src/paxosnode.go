package main

import (
	"consensuslib/paxosnode/paxosnodeinterface"
	"consensuslib"
)

type PaxosNodeInstance int

type Message = consensuslib.Message

var pn paxosnodeinterface.PaxosNode

// TODO[sharon]TODO[alex]: Implement rpcs
func main() {


	
}

// errors only happen for disconnections

func (pni *PaxosNodeInstance) ProcessPrepareRequest(m Message, r *Message) (err error) {
	*r = pn.Acceptor.ProcessPrepare(m)
	return nil
}

func (pni *PaxosNodeInstance) ProcessAcceptRequest(m Message, r *Message) (err error) {
	*r = pn.Acceptor.ProcessAccept(m)
	return nil
}

func (pni *PaxosNodeInstance) ProcessLearnRequest(m Message, r *Message) (err error) {
	// TODO: after Larissa's implementation put something like:
	// TODO: pn.Learner.Learn(m)
	*r = pn.Acceptor.ProcessAccept(m)
	if m.Equals(r) {
		// TODO: add func to call RPC "NotifyAboutAccepted"
	}
	return nil
}

// RPC call which is called by node that tries to connect
func (pni *PaxosNodeInstance) ConnectRemoteNeighbour (addr string, r *bool) (err error)  {
	err = pn.AcceptNeighbourConnection(addr, r)
	return err
}

// RPC call from other node's Acceptor about value it accepted
func (pni *PaxosNodeInstance) NotifyAboutAccepted (m * Message, r *bool) (err error) {
	pn.CountForNumAlreadyAccepted(m)
	return err
}
