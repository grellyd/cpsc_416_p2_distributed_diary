package paxosnode

import (
	"consensuslib/errors"
	"consensuslib/message"
	"consensuslib/paxosnode/acceptor"
	"consensuslib/paxosnode/learner"
	"consensuslib/paxosnode/proposer"
	"filelogger/singletonlogger"
	"paxostracker"
	"fmt"
	"net/rpc"
	"time"
	"regexp"
	"sync"
	"math/rand"
)

// Type Aliases
type ProposerRole = proposer.ProposerRole
type AcceptorRole = acceptor.AcceptorRole
type LearnerRole = learner.LearnerRole

var portRegex = regexp.MustCompile(":([0-9])+")

const TIMER  = 5*time.Second
// number to pick from 0 to RANDOFFSET to sleep between next proposals
const RANDOFFSET = 3
// TTL for message
const TTL = 3

type PaxosNode struct {
	Addr       string // IP:port, identifier
	Proposer   ProposerRole
	Acceptor   AcceptorRole
	Learner    LearnerRole
	NbrAddrs   []string
	Neighbours map[string]*rpc.Client
	FailedNeighbours []string
	RoundNum	 int
}

// A client will call this to mount to create a Paxos Node that
// is linked to the client. The PN's Addr field is set as the pnAddr passed in
func NewPaxosNode(pnAddr string) (pn *PaxosNode, err error) {
	proposer := proposer.NewProposer(pnAddr)
	acceptor := acceptor.NewAcceptor()
	learner := learner.NewLearner()
	pn = &PaxosNode{
		Addr:     pnAddr,
		Proposer: proposer,
		Acceptor: acceptor,
		Learner:  learner,
	}
	//portNumber := portRegex.FindString(pn.Addr)
	//acceptor.RestoreFromBackup(portNumber[1:])
	acceptor.RestoreFromBackup()
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] after backup restoration promised value is %v", acceptor.LastPromised))
	fmt.Sprintf("[paxosnode] after backup restoration promised value is %v", acceptor.LastPromised)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] after backup restoration accepted value is %v", acceptor.LastAccepted))
	fmt.Sprintf("[paxosnode] after backup restoration accepted value is %v", acceptor.LastAccepted)
	return pn, err
}

// TODO: Rename BecomeNeigbhors to SendNeighbors
func (pn *PaxosNode) SendNeighbours(ips []string) (err error) {
	err = pn.BecomeNeighbours(ips)
	return err
}

func (pn *PaxosNode) LearnLatestValueFromNeighbours() (err error) {
	err = pn.SetInitialLog()
	return err
}

func (pn *PaxosNode) UnmountPaxosNode() (err error) {
	// Close all RPC connections with neighbours during unmount
	for _, conn := range pn.Neighbours {
		conn.Close()
	}
	pn.NbrAddrs = nil

	return nil
}

// Handles the entire process of proposing a value and trying to achieve consensus
func (pn *PaxosNode) WriteToPaxosNode(value, msgHash string, ttl int) (success bool, err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Writing to paxos %v TTL: %v", value, ttl))
	fmt.Sprintf("[paxosnode] Writing to paxos %v", value)
	prepReq := pn.Proposer.CreatePrepareRequest(pn.RoundNum, msgHash, ttl)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Prepare request is id: %d , val: %s, type: %d, round: %d \n", prepReq.ID, prepReq.Value, prepReq.Type, prepReq.RoundNum))
	fmt.Sprintf("[paxosnode] Prepare request is id: %d , val: %s, type: %d, round: %d \n", prepReq.ID, prepReq.Value, prepReq.Type, prepReq.RoundNum)
	numAccepted, err := pn.DisseminateRequest(prepReq)
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Pledged to accept %v", numAccepted))
	// TODO: Unsure if err from DisseminateRequest should bubble up to client. Previous Note: should return new value?
	if err != nil {
		return false, err
	}

	// If majority is not reached, sleep for a while and try again
	// TODO: check whether should retry must return an error if no connection or something
	//pn.ShouldRetry(numAccepted, value, msgHash, prepReq.Bounces)
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
	fmt.Sprintf("[paxosnode] Accepted %v", numAccepted)
	// If majority is not reached, sleep for a while and try again
	//pn.ShouldRetry(numAccepted, value, msgHash, accReq.Bounces)
	b, e = pn.ShouldRetry(numAccepted,value, &accReq)
	if b {
		return b, e
	}
	paxostracker.Learn(uint64(numAccepted))

	// Remove all the failed neighbours at the end of a round
	// TODO: commented out since we consider paxos NW alive unless the majority failed
	//pn.ClearFailedNeighbours()
	paxostracker.Idle(value)

	return true, nil
}

// Sets up bidirectional RPC with all neighbours. Neighbours list is passed to the
// Paxos Node by the client.
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
		/*singletonlogger.Debug("[paxosnode] after I connected to PaxosNW length", len(pn.Neighbours), " and ngbours ")
		for k, v := range pn.Neighbours {
			singletonlogger.Debug(k, "and rpc ", v )
		}*/
	}
	return nil
}

// When a new node joins the network, it contacts all of its neighbours for their logs.
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
		if (len(temp) > maxLen) {
			maxLen = len(temp)
			longestLog = temp
		}
	}
	pn.Learner.InitializeLog(longestLog)

	// setting new messageId to a newly joined node to accommodate the same PSN across PaxosNW
	logLen := len(longestLog)
	if logLen != 0 {
		newMsgID := longestLog[len(longestLog)-1].ID
		pn.Proposer.UpdateMessageID(newMsgID)
	}

	return nil
}

func (pn *PaxosNode) SetRoundNum(roundNum int) {
	pn.RoundNum = roundNum
}

func (pn *PaxosNode) GetLog() (log []Message, err error) {
	log, err = pn.Learner.GetCurrentLog()
	return log, err
}

// This method sets up the bi-directional RPC. A new PN joins the network and will
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

// Disseminates a message to all neighbours. This includes prepare and accept requests.
func (pn *PaxosNode) DisseminateRequest(prepReq Message) (numAccepted int, err error) {
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] Disseminate request %v", prepReq.Type))
	numAccepted = 0
	//respReq := prepReq
	//respReq := *(new(Message))
	switch prepReq.Type {
	case message.PREPARE:
		singletonlogger.Debug("[paxosnode] PREPARE")

		// Set up timer and channel for responses
		//timer := time.NewTimer(time.Minute)
		timer := time.NewTimer(TIMER)
		defer timer.Stop()
		go func() {
			<- timer.C
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


		/*for k, v := range pn.Neighbours {

			singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating to neighbour %v", k))

			go func(v *rpc.Client, k string) {
				var respReq Message
				singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating to neighbour inside %v and RPC %v", k, v))
				errQueue <- v.Call("PaxosNodeRPCWrapper.ProcessPrepareRequest", prepReq, &respReq)
				c<-respReq
			}(v, k)
			select {
			case err := <- errQueue:
				singletonlogger.Debug("[paxosnode] channel worked on PREPARE")
				if err != nil {
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
					*//*if len(errQueue) >= len(pn.Neighbours)/2 {
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] checking errQueue %v", len(errQueue)))
						pn.RoundNum++
						return numAccepted, nil
					}*//*
					if len(pn.FailedNeighbours) >= len(pn.Neighbours)/2 {
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] checking failed nbrs %v", len(pn.FailedNeighbours)))
						pn.RoundNum++
						return numAccepted, nil
					}
					singletonlogger.Debug(fmt.Sprintf("[paxosnode] on PREPARE RPC failed %v", k))
				} else {
					req := <- c
					if prepReq.Equals(&req) {
						numAccepted++
						if pn.IsMajority(numAccepted) {
							//return numAccepted, nil
						}
					}
				}
			case <- time.After(TIMER):
				pn.FailedNeighbours = append(pn.FailedNeighbours, k)
			}

		}*/

		for k, v := range pn.Neighbours {

			singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating to neighbour %v", k))

			go func(v *rpc.Client, k string) {
				//wg.Add(1)
				defer wg.Done()
				var respReq Message
				singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating to neighbour inside %v and RPC %v", k, v))
				errQueue <- v.Call("PaxosNodeRPCWrapper.ProcessPrepareRequest", prepReq, &respReq)
				c<-respReq
				select {
				case err := <- errQueue:
					singletonlogger.Debug("[paxosnode] channel worked on PREPARE")
					if err != nil {
						pn.FailedNeighbours = append(pn.FailedNeighbours, k)
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] on PREPARE RPC failed %v", k))
					} else {
						req := <- c
						if prepReq.Equals(&req) {
							numAccepted++
							singletonlogger.Debug(fmt.Sprintf("[paxosnode] on PREPARE RPC succeded %v numPledged: %v, ID: %v", req.FromProposerID, numAccepted, req.ID))
						}
					}
				case <- time.After(TIMER):
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
				}
				//wg.Done()
			}(v, k)


		}
		wg.Wait()
		if len(pn.FailedNeighbours) >= len(pn.Neighbours)/2 && len(pn.FailedNeighbours) != 0{
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] checking failed nbrs %v", len(pn.FailedNeighbours)))
			//pn.RoundNum++
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

		/*for k, v := range pn.Neighbours {

			go func(k string, v *rpc.Client) {
				singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating ACCEPT to neighbour %v", k))
				errQueue <- v.Call("PaxosNodeRPCWrapper.ProcessAcceptRequest", prepReq, &respReq)
				c<-respReq

			}(k, v)
			select {
			case err := <- errQueue:
				singletonlogger.Debug("[paxosnode] channel worked on PREPARE")
				if err != nil {
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
					*//*if len(errQueue) >= len(pn.Neighbours)/2 {
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] checking errQueue %v", len(errQueue)))
						pn.RoundNum++
						return numAccepted, nil
					}*//*
					if len(pn.FailedNeighbours) >= len(pn.Neighbours)/2 {
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] checking failed nbrs %v", len(pn.FailedNeighbours)))
						pn.RoundNum++
						return numAccepted, nil
					}
					singletonlogger.Debug(fmt.Sprintf("[paxosnode] on PREPARE RPC failed %v", k))
				} else {
					req := <- c
					if prepReq.Equals(&req) {
						numAccepted++
					}
				}
			case <- time.After(TIMER):
				pn.FailedNeighbours = append(pn.FailedNeighbours, k)
			}


		}*/

		// last send it to ourselves
		//TODO: we accept first?
		resp := pn.Acceptor.ProcessAccept(prepReq, pn.RoundNum)
		if resp.Equals(&prepReq) {
			numAccepted++
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] I accepted and the # is %v", numAccepted))
			//go pn.SayAccepted(&prepReq)
			pn.SayAccepted(&prepReq)
		}

		for k, v := range pn.Neighbours {

			go func(k string, v *rpc.Client) {
				//wg.Add(1)
				defer wg.Done()
				var respReq Message
				singletonlogger.Debug(fmt.Sprintf("[paxosnode] disseminating ACCEPT to neighbour %v", k))
				errQueue <- v.Call("PaxosNodeRPCWrapper.ProcessAcceptRequest", prepReq, &respReq)
				c<-respReq
				select {
				case err := <- errQueue:
					singletonlogger.Debug("[paxosnode] channel worked on ACCEPT")
					if err != nil {
						pn.FailedNeighbours = append(pn.FailedNeighbours, k)
						singletonlogger.Debug(fmt.Sprintf("[paxosnode] on ACCEPT RPC failed %v", k))
					} else {
						req := <- c
						if prepReq.Equals(&req) {
							numAccepted++
							singletonlogger.Debug(fmt.Sprintf("[paxosnode] on ACCEPT RPC succeded %v numAccepted: %vID: %v", req.FromProposerID, numAccepted, req.ID))
						}
					}
				case <- time.After(TIMER):
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
				}
				//wg.Done()
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
	return numAccepted, err
}

// Notifies all learners that request was accepted
/*func (pn *PaxosNode) SayAccepted(m *Message) {
	// first, tell to own learner
	pn.CountForNumAlreadyAccepted(m)
	// then to all other nodes' learners
	var counted bool
	for k, v := range pn.Neighbours {
		e := v.Call("PaxosNodeRPCWrapper.NotifyAboutAccepted", m, &counted)
		if e != nil {
			pn.FailedNeighbours = append(pn.FailedNeighbours, k)
		}
	}
}*/

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

func (pn *PaxosNode) IsMajority(n int) bool {
	if n > (len(pn.Neighbours)+1)/2 {
		return true
	}
	return false
}

// This method takes role of Learner, adds Accepted message to the map of accepted messages,
// and notifies learner when the # for this particular message is a majority to write into the log
// TODO: think about moving this responsibility to the learner
func (pn *PaxosNode) CountForNumAlreadyAccepted(m *Message) {
	//singletonlogger.Debug("[paxosnode] in CountForNumAlreadyAccepted, round # ", pn.RoundNum)
	numSeen := pn.Learner.NumAlreadyAccepted(m)
	//singletonlogger.Debug("[paxosnode] in CountForNumAlreadyAccepted, how many accepted ", numSeen)
	if pn.IsMajority(numSeen) {
		// TODO: Learner.LearnValue returns the next round #; use the new round # somewhere?
		/*if !pn.IsInLog(m) {
			pn.Learner.LearnValue(m)
			pn.RoundNum++
		}*/
		pn.RoundNum, _ = pn.Learner.LearnValue(m)
		//pn.RoundNum++
		singletonlogger.Debug(fmt.Sprintf("[paxosnode] in CountForNumAlreadyAccepted, value learned, next round # %v", pn.RoundNum))
	}

}

func (pn *PaxosNode) ShouldRetry(numAccepted int, value string, m *Message) (b bool, err error) {
	if !pn.IsMajority(numAccepted) {
		singletonlogger.Debug("[paxosnode] We're retrying")
		m.Bounces--
		if m.Bounces == 0 {
			randOffset := time.Duration(rand.Intn(RANDOFFSET))
			singletonlogger.Debug(fmt.Sprintf("[paxosnode] sleeping for %v", randOffset))
			time.Sleep(randOffset*time.Second)
			m.Bounces = TTL
		}
		// Before retrying, we must clear the failed neighbours
		pn.ClearFailedNeighbours()
		pn.NotifyOfMajorityFailure()
		numAccepted = 0
		//time.Sleep(message.SLEEPTIME)
		b, err = pn.WriteToPaxosNode(value, m.MsgHash, m.Bounces)
	}
	return b, err
}

func (pn *PaxosNode) ClearFailedNeighbours() {
	for _, ip := range pn.FailedNeighbours {
		pn.RemoveFailedNeighbour(ip)
	}
	pn.FailedNeighbours = nil
	pn.RoundNum++
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] cleaned nbrs, new round is # %v", pn.RoundNum))
}

func (pn *PaxosNode) RemoveFailedNeighbour(ip string) {
	delete(pn.Neighbours, ip)
	pn.RemoveNbrAddr(ip)
}

func (pn *PaxosNode) RemoveNbrAddr(ip string) {
	for i, v := range pn.NbrAddrs {
		if v == ip {
			pn.NbrAddrs = append(pn.NbrAddrs[:i], pn.NbrAddrs[i+1:]...)
			break
		}
	}
}

func (pn *PaxosNode) NotifyOfMajorityFailure () {
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
			case err := <- errQueue:
				singletonlogger.Debug("[paxosnode] channel worked on MAJOR FAILURE")
				if err != nil {
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
					singletonlogger.Debug(fmt.Sprintf("[paxosnode] on MAJOR FAILURE RPC failed %v", k))
				}
			case <- time.After(TIMER):
				pn.FailedNeighbours = append(pn.FailedNeighbours, k)
			}
		}(k, v)
	}
	wg.Wait()
	//pn.ClearFailedNeighbours()
	singletonlogger.Debug(fmt.Sprintf("[paxosnode] notified nbrs, new round is # %v", pn.RoundNum))
}

func (pn *PaxosNode) CleanNbrsOnRequest(neighbour string) (b bool) {
	nghbrNum := len(pn.Neighbours)-1
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
			case err := <- errQueue:
				singletonlogger.Debug("[paxosnode] channel worked on CLEANING")
				if err != nil {
					pn.FailedNeighbours = append(pn.FailedNeighbours, k)
					singletonlogger.Debug(fmt.Sprintf("[paxosnode] on CLEANING failed %v", k))
				}
			case <- time.After(TIMER):
				pn.FailedNeighbours = append(pn.FailedNeighbours, k)
			}
		}(k, v)
	}
	wg.Wait()
	pn.ClearFailedNeighbours()
	return true
}

/*func (pn *PaxosNode) IsInLog(m *Message) bool {
	for _,v := range pn.Learner.Log {
		if v.MsgHash == m.MsgHash {
			return true
		}
	}
	return false
}*/
