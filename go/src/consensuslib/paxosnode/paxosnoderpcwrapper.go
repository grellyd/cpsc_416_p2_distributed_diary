package paxosnode

import (
	"filelogger/singletonlogger"
	"consensuslib/message"
	"fmt"
	//"sync"
	//"net/rpc"
	//"time"
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

// RPCs for paxosnodes start here
func (p *PaxosNodeRPCWrapper) ProcessPrepareRequest(m Message, r *Message) (err error) {
	singletonlogger.Debug("[paxosnodewrapper] increasing message ID")
	p.paxosNode.Proposer.IncrementMessageID()
	*r = p.paxosNode.Acceptor.ProcessPrepare(m, p.paxosNode.RoundNum)
	return nil
}

// RPC call received from other node to process accept request
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

// *****
/*func (p *PaxosNodeRPCWrapper) ProcessLearnRequest(m Message, r *Message) (err error) {
	p.paxosNode.Learner.LearnValue(&m) // TODO: We don't consider round numbers or indices
	*r = p.paxosNode.Acceptor.ProcessAccept(m)
	return nil
}*/

// RPC call which is called by node that tries to connect
func (p *PaxosNodeRPCWrapper) ConnectRemoteNeighbour(addr string, r *bool) (err error) {
	//singletonlogger.Debug("[paxoswrapper] connecting my remote neighbour")
	err = p.paxosNode.AcceptNeighbourConnection(addr, r)
	//singletonlogger.Debug("[paxoswrapper] error on connection? ", *r)
	return err
}

// RPC call from other node's Acceptor about value it accepted
func (p *PaxosNodeRPCWrapper) NotifyAboutAccepted(m *Message, r *bool) (err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnodewrapper] notify about accepted %v", m.Type))
	p.paxosNode.CountForNumAlreadyAccepted(m)
	return err
}

// RPC call from a new PN that joined the network and needs to read
// the state of the log from every other PN's learner
func (p *PaxosNodeRPCWrapper) ReadFromLearner(placeholder string, log *[]Message) (err error) {
	*log, err = p.paxosNode.GetLog()
	// return no errors, for now
	return nil
}

func (p *PaxosNodeRPCWrapper) CleanYourNeighbours (neighbour string, b *bool) (err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnodewrapper] cleaning request from %s", neighbour))
	*b = p.paxosNode.CleanNbrsOnRequest(neighbour)
	return nil
}

func (p *PaxosNodeRPCWrapper) RUAlive (placeholder string, b *bool) (err error) {
	*b = true
	return nil
}

/* Unused for now
func (p *PaxosNodeRPCWrapper) GetLastPromisedProposal(placeholder string, proposal *Message) (err error) {
	*proposal = p.paxosNode.Acceptor.LastPromised
	return nil
}
*/
