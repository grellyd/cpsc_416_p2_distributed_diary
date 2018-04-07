package paxosnode

import (
	"consensuslib/errors"
	"consensuslib/message"
	"consensuslib/paxosnode/acceptor"
	"consensuslib/paxosnode/learner"
	"consensuslib/paxosnode/proposer"
	"filelogger/singletonlogger"
	"fmt"
	"math/rand"
	"net/rpc"
	"paxostracker"
	"regexp"
	"sync"
	"time"
)

/**
 * PaxosNode implements the interface that the rest of the consensuslib talks to.
 * It in turns make calls to internal the Learner, Acceptor, and Proposer roles.
 *
 *	See paxosnodeinterface.go for the public methods that it implements and their descriptions.
 */

// ProposerRole Type Alias
type ProposerRole = proposer.ProposerRole

// AcceptorRole Type Alias
type AcceptorRole = acceptor.AcceptorRole

// LearnerRole Type Alias
type LearnerRole = learner.LearnerRole

var portRegex = regexp.MustCompile(":([0-9])+")

// TIMER for timeouts
const TIMER = 5 * time.Second

// RANDOFFSET where pick a number from 0 to RANDOFFSET to sleep between next proposals
const RANDOFFSET = 3

// TTL for message
const TTL = 3

// PaxosNode struct
type PaxosNode struct {
	Addr             string // IP:port, identifier
	Proposer         ProposerRole
	Acceptor         AcceptorRole
	Learner          LearnerRole
	NbrAddrs         []string
	Neighbours       map[string]*rpc.Client
	FailedNeighbours []string
	RoundNum         int
}

// NewPaxosNode creates a Paxos Node that is linked to the client. The PN's Addr field is set as the pnAddr passed in
func NewPaxosNode(pnAddr string) (pn *PaxosNode, err error) {
	proposer := proposer.NewProposer(pnAddr)
	acceptorID := portRegex.FindString(pnAddr)
	acceptor := acceptor.NewAcceptor(acceptorID)
	learner := learner.NewLearner()
	pn = &PaxosNode{
		Addr:     pnAddr,
		Proposer: proposer,
		Acceptor: acceptor,
		Learner:  learner,
	}
	acceptor.RestoreFromBackup()
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] after backup restoration promised value is %v", acceptor.LastPromised))
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] after backup restoration accepted value is %v", acceptor.LastAccepted))
	return pn, err
}

// LearnLatestValueFromNeighbours is for the inital setup
func (pn *PaxosNode) LearnLatestValueFromNeighbours() (err error) {
	err = pn.SetInitialLog()
	return err
}

// UnmountPaxosNode closes all RPC connections with neighbours nicely
func (pn *PaxosNode) UnmountPaxosNode() (err error) {
	for _, conn := range pn.Neighbours {
		conn.Close()
	}
	pn.NbrAddrs = nil

	return nil
}

// WriteToPaxosNode Handles the entire process of proposing a value and trying to achieve consensus
func (pn *PaxosNode) WriteToPaxosNode(value, msgHash string, ttl int) (success bool, err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Writing to paxos %v TTL: %v", value, ttl))
	prepReq := pn.Proposer.CreatePrepareRequest(pn.RoundNum, msgHash, ttl)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Prepare request is id: %d , val: %s, type: %d, round: %d \n", prepReq.ID, prepReq.Value, prepReq.Type, prepReq.RoundNum))
	numAccepted, err := pn.DisseminateRequest(prepReq)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Pledged to accept %v", numAccepted))
	if err != nil {
		singletonlogger.Error(err.Error())
		return false, err
	}

	// If majority is not reached, sleep for a while and try again
	b, e := pn.ShouldRetry(numAccepted, value, &prepReq)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] returned from should retry positively %v \n", b))
	if b {
		return b, e
	}

	accReq := pn.Proposer.CreateAcceptRequest(value, msgHash, pn.RoundNum, prepReq.Bounces)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Accept request is id: %d , val: %s, type: %d \n", accReq.ID, accReq.Value, accReq.Type))
	paxostracker.Propose(accReq.ID)
	numAccepted, err = pn.DisseminateRequest(accReq)
	if err != nil {
		return false, err
	}
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Accepted %v", numAccepted))
	// If majority is not reached, sleep for a while and try again
	b, e = pn.ShouldRetry(numAccepted, value, &accReq)
	if b {
		return b, e
	}

	return true, nil
}

// BecomeNeighbours sets up bidirectional RPC with all neighbours
func (pn *PaxosNode) BecomeNeighbours(ips []string) (err error) {
	for _, ip := range ips {
		neighbourConn, err := rpc.Dial("tcp", ip)
		if err != nil {
			singletonlogger.Debug("[paxosnode]: Error in BecomeNeighbours")
			return errors.NeighbourConnectionError(ip)
		}
		connected := false
		err = neighbourConn.Call("PaxosNodeRPCWrapper.ConnectRemoteNeighbour", pn.Addr, &connected)
		// Add ip to connectedNbrs and add the connection to Neighbours map
		// after bidirectional RPC connection establishment is successful
		if connected {
			singletonlogger.Debug("[paxosnode]: connected to the nbr")
			pn.NbrAddrs = append(pn.NbrAddrs, ip)
			if pn.Neighbours == nil {
				pn.Neighbours = make(map[string]*rpc.Client, 0)
			}
			pn.Neighbours[ip] = neighbourConn
		}
	}
	return nil
}

// SetInitialLog when a new node joins the network by contacting all of its neighbours for their logs.
// The new node will then set its initial log to be the longest log received from neighbours
func (pn *PaxosNode) SetInitialLog() (err error) {
	singletonlogger.Debug("[paxosnode] Setting the initial log for this new node")
	maxLen := 0
	longestLog := make([]Message, 0)
	for k, v := range pn.Neighbours {
		// Create a temporary log to get filled by neighbour learners
		temp := make([]Message, 0)
		singletonlogger.Debug(fmt.Sprintf("[paxosnode] Making ReadFromLearner call to node %v\n", v))
		e := v.Call("PaxosNodeRPCWrapper.ReadFromLearner", "placeholder", &temp)
		if e != nil {
			pn.RemoveFailedNeighbour(k)
			continue
		}
		if len(temp) > maxLen {
			maxLen = len(temp)
			longestLog = temp
		}
	}
	pn.Learner.InitializeLog(longestLog)

	// Set a new messageId to a newly joined node to accommodate the same PSN across PaxosNW
	logLen := len(longestLog)
	if logLen != 0 {
		newMsgID := longestLog[len(longestLog)-1].ID
		pn.Proposer.UpdateMessageID(newMsgID)
	}

	return nil
}

// SetRoundNum helper method
func (pn *PaxosNode) SetRoundNum(roundNum int) {
	pn.RoundNum = roundNum
}

// GetLog of the pn's learner
func (pn *PaxosNode) GetLog() (log []Message, err error) {
	log, err = pn.Learner.GetCurrentLog()
	return log, err
}

// AcceptNeighbourConnection sets up the bi-directional RPC. A new PN joins the network and will
// establish an RPC connection with each of the other PNs
func (pn *PaxosNode) AcceptNeighbourConnection(addr string, result *bool) (err error) {
	neighbourConn, err := rpc.Dial("tcp", addr)
	if err != nil {
		singletonlogger.Debug("[paxosnode] Error in AcceptNeighbourConnection")
		return errors.NeighbourConnectionError(addr)
	}
	pn.NbrAddrs = append(pn.NbrAddrs, addr)
	if pn.Neighbours == nil {
		pn.Neighbours = make(map[string]*rpc.Client, 0)
	}
	pn.Neighbours[addr] = neighbourConn

	neighbors := ""
	for _, n := range pn.Neighbours {
		neighbors += fmt.Sprintf("%v ", n)
	}
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] after neigh connection we have length '%v' and neighbours %v", len(pn.Neighbours), neighbors))
	*result = true
	return nil
}

// DisseminateRequest sends a message to all neighbours. This includes prepare and accept requests.
func (pn *PaxosNode) DisseminateRequest(prepReq Message) (numAccepted int, err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Disseminate request %v", prepReq.Type))
	numAccepted = 0
	switch prepReq.Type {
	case message.PREPARE:
		singletonlogger.Debug("[paxosnode] PREPARE")

		// Set up timer and channel for responses
		timer := time.NewTimer(TIMER)
		defer timer.Stop()
		go func() {
			<-timer.C
		}()

		nghbrNum := len(pn.Neighbours)
		c := make(chan Message, nghbrNum)
		errQueue := make(chan error, nghbrNum)
		var wg sync.WaitGroup
		wg.Add(nghbrNum)

		// first send it to ourselves
		resp := pn.Acceptor.ProcessPrepare(prepReq, pn.RoundNum)
		if resp.Equals(&prepReq) {
			numAccepted++
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] I pledged and the # is %v", numAccepted))
		}

		for k, v := range pn.Neighbours {

			singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating to neighbour %v", k))

			go func(v *rpc.Client, k string) {
				defer wg.Done()
				var respReq Message
				singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating to neighbour inside %v and RPC %v", k, v))
				errQueue <- v.Call("PaxosNodeRPCWrapper.ProcessPrepareRequest", prepReq, &respReq)
				c <- respReq
				select {
				case err := <-errQueue:
					singletonlogger.Debug("[paxosnode] channel worked on PREPARE")
					if err != nil {
						pn.FailedNeighbours = append(pn.FailedNeighbours, k)
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] on PREPARE RPC failed %v", k))
					} else {
						req := <-c
						if prepReq.Equals(&req) {
							numAccepted++
							singletonlogger.Debug(fmt.Sprintf("[paxosnode] on PREPARE RPC succeded %v numPledged: %v, ID: %v", req.FromProposerID, numAccepted, req.ID))
						}
					}
				case <-time.After(TIMER):
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
				}
			}(v, k)

		}
		wg.Wait()
		if len(pn.FailedNeighbours) >= len(pn.Neighbours)/2 && len(pn.FailedNeighbours) != 0 {
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] checking failed nbrs %v", len(pn.FailedNeighbours)))
			return numAccepted, nil
		}

		return numAccepted, nil

	case message.ACCEPT:
		singletonlogger.Debug("[paxosnode] ACCEPT")
		nghbrNum := len(pn.Neighbours)
		c := make(chan Message, nghbrNum)
		errQueue := make(chan error, nghbrNum)
		var wg sync.WaitGroup
		wg.Add(nghbrNum)

		// last send it to ourselves
		resp := pn.Acceptor.ProcessAccept(prepReq, pn.RoundNum)
		if resp.Equals(&prepReq) {
			numAccepted++
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] I accepted and the # is %v", numAccepted))
			pn.SayAccepted(&prepReq)
		}

		for k, v := range pn.Neighbours {

			go func(k string, v *rpc.Client) {
				defer wg.Done()
				var respReq Message
				singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating ACCEPT to neighbour %v", k))
				errQueue <- v.Call("PaxosNodeRPCWrapper.ProcessAcceptRequest", prepReq, &respReq)
				c <- respReq
				select {
				case err := <-errQueue:
					singletonlogger.Debug("[paxosnode] channel worked on ACCEPT")
					if err != nil {
						pn.FailedNeighbours = append(pn.FailedNeighbours, k)
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] on ACCEPT RPC failed %v", k))
					} else {
						req := <-c
						if prepReq.Equals(&req) {
							numAccepted++
							singletonlogger.Debug(fmt.Sprintf("[paxosnode] on ACCEPT RPC succeded %v numAccepted: %vID: %v", req.FromProposerID, numAccepted, req.ID))
						}
					}
				case <-time.After(TIMER):
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
				}
			}(k, v)
		}

		wg.Wait()

		if len(pn.FailedNeighbours) >= len(pn.Neighbours)/2 && len(pn.FailedNeighbours) != 0 {
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] checking failed nbrs %v", len(pn.FailedNeighbours)))
			pn.RoundNum++
			return numAccepted, nil
		}

		return numAccepted, nil

	default:
		return -1, errors.InvalidMessageTypeError(prepReq)
	}
}

// SayAccepted sends an accept message
func (pn *PaxosNode) SayAccepted(m *Message) {
	// first, tell to own learner
	pn.CountForNumAlreadyAccepted(m)
	// then to all other nodes' learners

	for k, v := range pn.Neighbours {
		go func(k string, v *rpc.Client) {
			var counted bool
			e := v.Call("PaxosNodeRPCWrapper.NotifyAboutAccepted", m, &counted)
			if e != nil {
				pn.FailedNeighbours = append(pn.FailedNeighbours, k)
			}
		}(k, v)

	}
}

// IsMajority helper method
func (pn *PaxosNode) IsMajority(n int) bool {
	if n > (len(pn.Neighbours)+1)/2 {
		return true
	}
	return false
}

// CountForNumAlreadyAccepted takes role of Learner, adds Accepted message to the map of accepted messages,
// and notifies learner when the # for this particular message is a majority to write into the log
func (pn *PaxosNode) CountForNumAlreadyAccepted(m *Message) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] in CountForNumAlreadyAccepted, round # %v", pn.RoundNum))
	numSeen := pn.Learner.NumAlreadyAccepted(m)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] in CountForNumAlreadyAccepted, how many accepted %v", numSeen))
	if pn.IsMajority(numSeen) {
		pn.RoundNum, _ = pn.Learner.LearnValue(m)
		singletonlogger.Debug(fmt.Sprintf("[paxosnode] in CountForNumAlreadyAccepted, value learned, next round # %v", pn.RoundNum))
	}

}

// ShouldRetry checks if the round should be retried due to a lack of majority
func (pn *PaxosNode) ShouldRetry(numAccepted int, value string, m *Message) (b bool, err error) {
	if !pn.IsMajority(numAccepted) {
		singletonlogger.Debug("[paxosnode] We're retrying")
		m.Bounces--
		if m.Bounces == 0 {
			randOffset := time.Duration(rand.Intn(RANDOFFSET))
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] sleeping for %v", randOffset))
			time.Sleep(randOffset * time.Second)
			m.Bounces = TTL
		}
		// Before retrying, we must clear the failed neighbours
		pn.ClearFailedNeighbours()
		pn.NotifyOfMajorityFailure()
		numAccepted = 0
		b, err = pn.WriteToPaxosNode(value, m.MsgHash, m.Bounces)
	}
	return b, err
}

// ClearFailedNeighbours removes failed neighbors from a pn's collection
func (pn *PaxosNode) ClearFailedNeighbours() {
	for _, ip := range pn.FailedNeighbours {
		pn.RemoveFailedNeighbour(ip)
	}
	pn.FailedNeighbours = nil
	pn.RoundNum++
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] cleaned nbrs, new round is # %v", pn.RoundNum))
}

// RemoveFailedNeighbour removes a single neighbour
func (pn *PaxosNode) RemoveFailedNeighbour(ip string) {
	delete(pn.Neighbours, ip)
	pn.RemoveNbrAddr(ip)
}

// RemoveNbrAddr removes a Neighbour's addreess
func (pn *PaxosNode) RemoveNbrAddr(ip string) {
	for i, v := range pn.NbrAddrs {
		if v == ip {
			pn.NbrAddrs = append(pn.NbrAddrs[:i], pn.NbrAddrs[i+1:]...)
			break
		}
	}
}

// NotifyOfMajorityFailure helper
func (pn *PaxosNode) NotifyOfMajorityFailure() {
	nghbrNum := len(pn.Neighbours)
	var wg sync.WaitGroup
	wg.Add(nghbrNum)
	var b bool
	c := make(chan bool, nghbrNum)
	errQueue := make(chan error, nghbrNum)

	for k, v := range pn.Neighbours {
		go func(k string, v *rpc.Client) {
			defer wg.Done()
			errQueue <- v.Call("PaxosNodeRPCWrapper.CleanYourNeighbours", k, &b)
			c <- b

			select {
			case err := <-errQueue:
				singletonlogger.Debug("[paxosnode] channel worked on MAJOR FAILURE")
				if err != nil {
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
					singletonlogger.Debug(fmt.Sprintf("[paxosnode] on MAJOR FAILURE RPC failed %v", k))
				}
			case <-time.After(TIMER):
				pn.FailedNeighbours = append(pn.FailedNeighbours, k)
			}
		}(k, v)
	}
	wg.Wait()
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] notified nbrs, new round is # %v", pn.RoundNum))
}

// CleanNbrsOnRequest to remove neighbours when requested
func (pn *PaxosNode) CleanNbrsOnRequest(neighbour string) (b bool) {
	nghbrNum := len(pn.Neighbours) - 1
	var wg sync.WaitGroup
	wg.Add(nghbrNum)
	c := make(chan bool, nghbrNum)
	errQueue := make(chan error, nghbrNum)

	for k, v := range pn.Neighbours {
		if k == neighbour {
			continue
		}
		go func(k string, v *rpc.Client) {
			defer wg.Done()
			errQueue <- v.Call("PaxosNodeRPCWrapper.RUAlive", k, &b)
			c <- b
			select {
			case err := <-errQueue:
				singletonlogger.Debug("[paxosnode] channel worked on CLEANING")
				if err != nil {
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
					singletonlogger.Debug(fmt.Sprintf("[paxosnode] on CLEANING failed %v", k))
				}
			case <-time.After(TIMER):
				pn.FailedNeighbours = append(pn.FailedNeighbours, k)
			}
		}(k, v)
	}
	wg.Wait()
	pn.ClearFailedNeighbours()
	return true
}
