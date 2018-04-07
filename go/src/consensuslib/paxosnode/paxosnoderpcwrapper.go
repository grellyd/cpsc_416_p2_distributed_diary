// this class represents the node wrapper which allows to make RPC calls
// between the Paxos Nodes (PNs)

package paxosnode

import (
	"consensuslib/message"
	"filelogger/singletonlogger"
	"fmt"
)

type Message = message.Message

type PaxosNodeRPCWrapper struct {
	paxosNode *PaxosNode
}

func NewPaxosNodeRPCWrapper(paxosNode *PaxosNode) (wrapper *PaxosNodeRPCWrapper, err error) {
	wrapper = &PaxosNodeRPCWrapper{
		paxosNode: paxosNode,
	}
	return wrapper, nil
}

// RPC to a PN's acceptor to process a new Prepare Request
func (p *PaxosNodeRPCWrapper) ProcessPrepareRequest(m Message, r *Message) (err error) {
	singletonlogger.Debug("[paxosnodewrapper] increasing message ID")
	p.paxosNode.Proposer.IncrementMessageID()
	*r = p.paxosNode.Acceptor.ProcessPrepare(m, p.paxosNode.RoundNum)
	return nil
}

// RPC to a PN's acceptor to process a new Accept Request
// If the request accepted, it gets disseminated to all the Learners in the Paxos NW
func (p *PaxosNodeRPCWrapper) ProcessAcceptRequest(m Message, r *Message) (err error) {
	singletonlogger.Debug("[paxosnodewrapper] RPC processing accept request")
	*r = p.paxosNode.Acceptor.ProcessAccept(m, p.paxosNode.RoundNum)
	if m.Equals(r) {
		singletonlogger.Debug("[paxosnodewrapper] saying accepted")
		go p.paxosNode.SayAccepted(r)
	}
	return nil
}

// RPC which is called by another node that tries to connect to the current one
func (p *PaxosNodeRPCWrapper) ConnectRemoteNeighbour(addr string, r *bool) (err error) {
	singletonlogger.Debug("[paxoswrapper] connecting my remote neighbour")
	err = p.paxosNode.AcceptNeighbourConnection(addr, r)
	//singletonlogger.Debug("[paxoswrapper] error on connection? ", *r)
	return err
}

// RPC to the Learner from other node's Acceptor about value it accepted
func (p *PaxosNodeRPCWrapper) NotifyAboutAccepted(m *Message, r *bool) (err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnodewrapper] notify about accepted %v", m.Type))
	p.paxosNode.CountForNumAlreadyAccepted(m)
	return err
}

// RPC from a new PN that joined the network and needs to read
// the state of the log from every other PN's learner
func (p *PaxosNodeRPCWrapper) ReadFromLearner(placeholder string, log *[]Message) (err error) {
	*log, err = p.paxosNode.GetLog()
	return nil
}

// RPC to notify a PN that majority failed and needs to be recalibrated
// makes a call to a node to clean failed neighbours
func (p *PaxosNodeRPCWrapper) CleanYourNeighbours(neighbour string, b *bool) (err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnodewrapper] cleaning request from %s", neighbour))
	*b = p.paxosNode.CleanNbrsOnRequest(neighbour)
	return nil
}

// RPC that asks a PN whether it still alive
func (p *PaxosNodeRPCWrapper) RUAlive(placeholder string, b *bool) (err error) {
	*b = true
	return nil
}
