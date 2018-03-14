package main

import (
	"consensuslib/paxosnode/paxosnodeinterface"
	"consensuslib"
)

type PaxosNodeInstance int

type Message = consensuslib.Message

var pn paxosnodeinterface.PaxosNode

func main() {
	
}

// errors only happen for disconnections
func (pni *PaxosNodeInstance) ReadPrepareRequest(m Message, r *Message) (err error) {
	*r = pn.Acceptor.ProcessAccept(m)
	return nil
}