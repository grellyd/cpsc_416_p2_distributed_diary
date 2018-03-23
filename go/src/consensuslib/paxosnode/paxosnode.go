package paxosnode

import (
	"consensuslib/errors"
	"consensuslib/message"
	"consensuslib/paxosnode/acceptor"
	"consensuslib/paxosnode/learner"
	"consensuslib/paxosnode/proposer"
	"fmt"
	"net/rpc"
	"time"
	"regexp"
)

// Type Aliases
type ProposerRole = proposer.ProposerRole
type AcceptorRole = acceptor.AcceptorRole
type LearnerRole = learner.LearnerRole

var portRegex = regexp.MustCompile(":([0-9])+")

type PaxosNode struct {
	Addr       string // IP:port, identifier
	Proposer   ProposerRole
	Acceptor   AcceptorRole
	Learner    LearnerRole
	NbrAddrs   []string
	Neighbours map[string]*rpc.Client
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
	portNumber := portRegex.FindString(pn.Addr)
	acceptor.RestoreFromBackup(portNumber[1:])
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
//TODO[sharon]: update parameters as needed.
func (pn *PaxosNode) WriteToPaxosNode(value string) (success bool, err error) {
	fmt.Println("[paxosnode] Writing to paxos ", value)
	prepReq := pn.Proposer.CreatePrepareRequest()
	fmt.Printf("[paxosnode] Prepare request is id: %d , val: %s, type: %d \n", prepReq.ID, prepReq.Value, prepReq.Type)
	numAccepted, err := pn.DisseminateRequest(prepReq)
	fmt.Println("[paxosnode] Pledged to accept ", numAccepted)
	// TODO: Unsure if err from DisseminateRequest should bubble up to client. Previous Note: should return new value?
	if err != nil {
		return false, err
	}

	// If majority is not reached, sleep for a while and try again
	// TODO: check whether should retry must return an error if no connection or something
	pn.ShouldRetry(numAccepted, value)

	// ***Unused For now***
	// Get the value of the highest-numbered proposal previously accepted among all acceptors, if any
	//previousProposedValue := pn.GetPreviousProposedValue()
	//if previousProposedValue != "" {
	//	value = previousProposedValue
	//}

	accReq := pn.Proposer.CreateAcceptRequest(value)
	fmt.Printf("[paxosnode] Accept request is id: %d , val: %s, type: %d \n", accReq.ID, accReq.Value, accReq.Type)
	numAccepted, err = pn.DisseminateRequest(accReq)
	if err != nil {
		return false, err
	}
	fmt.Println("[paxosnode] Accepted ", numAccepted)
	// If majority is not reached, sleep for a while and try again
	// TODO: check whether should retry must return an error if no connection or something
	pn.ShouldRetry(numAccepted, value)

	accReq.Type = message.CONSENSUS
	_, err = pn.DisseminateRequest(accReq)
	if err != nil {
		return false, err
	}

	return success, nil
}

// Sets up bidirectional RPC with all neighbours. Neighbours list is passed to the
// Paxos Node by the client.
func (pn *PaxosNode) BecomeNeighbours(ips []string) (err error) {
	for _, ip := range ips {
		neighbourConn, err := rpc.Dial("tcp", ip)
		if err != nil {
			return errors.NeighbourConnectionError(ip)
		}
		connected := false
		err = neighbourConn.Call("PaxosNodeRPCWrapper.ConnectRemoteNeighbour", pn.Addr, &connected)

		// Add ip to connectedNbrs and add the connection to Neighbours map
		// after bidirectional RPC connection establishment is successful
		if connected {
			fmt.Println("[paxosnode]: connected to the nbr")
			pn.NbrAddrs = append(pn.NbrAddrs, ip)
			if pn.Neighbours == nil {
				pn.Neighbours = make(map[string]*rpc.Client, 0)
			}
			pn.Neighbours[ip] = neighbourConn
		}
	}
	return nil
}

func (pn *PaxosNode) SetInitialLog() (err error) {
	logs := make(map[string]int, 0)
	for k, v := range pn.Neighbours {
		temp := make([]Message, 0)
		e := v.Call("PaxosNodeRPCWrapper.ReadFromLearner", "placeholder", &temp)
		if e != nil {
			pn.RemoveFailedNeighbour(k)
			continue
		}
		// Check if learners even have any messages written
		if (len(temp) > 0) {
			latestMsg := temp[len(temp)-1].Value
			if count, ok := logs[latestMsg]; ok {
				count++
				logs[latestMsg] = count
				// Once a majority is reached, set the initial log state to be the majority log
				if pn.IsMajority(count) {
					pn.Learner.InitializeLog(temp)
				}
			} else {
				logs[latestMsg] = 1
			}
		}
	}
	return nil
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
		return errors.NeighbourConnectionError(addr)
	}
	pn.NbrAddrs = append(pn.NbrAddrs, addr)
	if pn.Neighbours == nil {
		pn.Neighbours = make(map[string]*rpc.Client, 0)
	}
	pn.Neighbours[addr] = neighbourConn
	*result = true
	return nil
}

// Disseminates a message to all neighbours. This includes prepare and accept requests.
//TODO[sharon]: Figure out best name for number field and add as param. Might be RPC
func (pn *PaxosNode) DisseminateRequest(prepReq Message) (numAccepted int, err error) {
	fmt.Println("[paxosnode] Disseminate request")
	numAccepted = 0
	respReq := prepReq
	switch prepReq.Type {
	case message.PREPARE:
		fmt.Println("[paxosnode] PREPARE")
		for k, v := range pn.Neighbours {
			e := v.Call("PaxosNodeRPCWrapper.ProcessPrepareRequest", prepReq, &respReq)
			if e != nil {
				pn.RemoveFailedNeighbour(k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				if prepReq.Equals(&respReq) {
					numAccepted++
				}
			}
		}
		// last send it to ourselves
		pn.Acceptor.ProcessPrepare(prepReq)
		if prepReq.Equals(&respReq) {
			numAccepted++
		}
	case message.ACCEPT:
		fmt.Println("[paxosnode] ACCEPT")
		for k, v := range pn.Neighbours {
			e := v.Call("PaxosNodeRPCWrapper.ProcessAcceptRequest", prepReq, &respReq)
			if e != nil {
				pn.RemoveFailedNeighbour(k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				if prepReq.Equals(&respReq) {
					numAccepted++
				}
			}
		}
		// last send it to ourselves
		pn.Acceptor.ProcessAccept(prepReq)
		if prepReq.Equals(&respReq) {
			numAccepted++
			fmt.Println("[paxosnode] saying accepted for myself")
			go pn.SayAccepted(&prepReq)
		}
	case message.CONSENSUS:
		for k, v := range pn.Neighbours {
			e := v.Call("PaxosNodeRPCWrapper.ProcessLearnRequest", prepReq, &respReq)
			if e != nil {
				pn.RemoveFailedNeighbour(k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				if prepReq.Equals(&respReq) {
					numAccepted++
				}
			}
		}
	default:
		return -1, errors.InvalidMessageTypeError(prepReq)
	}

	return numAccepted, err
}

// Notifies all learners that request was accepted
func (pn *PaxosNode) SayAccepted(m *Message) {
	// first, tell to own learner
	pn.CountForNumAlreadyAccepted(m)
	// then to all other nodes' learners
	var counted bool
	for k, v := range pn.Neighbours {
		e := v.Call("PaxosNodeRPCWrapper.NotifyAboutAccepted", m, &counted)
		if e != nil {
			pn.RemoveFailedNeighbour(k)
		}
	}
}

// Locally accepts the accept request sent by a PN in the system.
// TODO[sharon]: Figure out parameters. Might be RPC
func AcceptAcceptRequest() (err error) {
	return err
}

func (pn *PaxosNode) IsMajority(n int) bool {
	if n > len(pn.Neighbours)/2 {
		return true
	}
	return false
}

// This method takes role of Learner, adds Accepted message to the map of accepted messages,
// and notifies learner when the # for this particular message is a majority to write into the log
// TODO: think about moving this responsibility to the learner
func (pn *PaxosNode) CountForNumAlreadyAccepted(m *Message) {
	fmt.Println("[paxosnode] in CountForNumAlreadyAccepted")
	numSeen := pn.Learner.NumAlreadyAccepted(m)
	if pn.IsMajority(numSeen) {
		pn.Learner.LearnValue(m) // this should write to the log TODO: expansion make learner return next round
	}
}

func (pn *PaxosNode) ShouldRetry(numAccepted int, value string) {
	if !pn.IsMajority(numAccepted) {
		time.Sleep(message.SLEEPTIME)
		pn.WriteToPaxosNode(value)
	}
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
/* Unused for now
func (pn *PaxosNode) GetPreviousProposedValue() string {
	highestProposal := uint64(0)
	priorProposedValue := ""
	// First check PN's neighbours to find the value of the highest-numbered proposal that they have accepted
	for k, v := range pn.Neighbours {
		var proposal Message
		e := v.Call("PaxosNodeRPCWrapper.GetLastPromisedProposal", "placeholder", &proposal)
		if e != nil {
			pn.RemoveFailedNeighbour(k)
		}

		// Check if proposal is not an empty message first (when neighbours have not accepted any proposals yet)
		// to avoid accessing fields of an empty struct
		if (Message{}) != proposal {
			if proposal.ID > highestProposal {
				highestProposal = proposal.ID
				priorProposedValue = proposal.Value
			}
		}
	}

	// Then check PN itself if it has already accepted a prior proposal
	selfLastPromisedProposal := pn.Acceptor.LastPromised
	if (Message{}) != selfLastPromisedProposal {
		if selfLastPromisedProposal.ID > highestProposal {
			priorProposedValue = selfLastPromisedProposal.Value
		}
	}

	return priorProposedValue
}
*/
