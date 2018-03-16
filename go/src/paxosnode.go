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
	return nil
}

// RPC call which is called by node that tries to connect
func (pni *PaxosNodeInstance) ConnectRemoteNeighbour (addr string, r *bool) (err error)  {
	err = pn.AcceptNeighbourConnection(addr, r)
	return err
}
