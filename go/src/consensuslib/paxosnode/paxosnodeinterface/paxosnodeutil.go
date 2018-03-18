package paxosnodeinterface

import (
	. "consensuslib"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

//type PaxosNodeInstance int

// Handles the entire process of proposing a value and trying to achieve consensus
//TODO[sharon]: update parameters as needed.
func (pn *PaxosNode) WriteToPaxosNode(value string) (success bool, err error) {
	prepReq := pn.Proposer.CreatePrepareRequest()
	numAccepted, err := pn.DisseminateRequest(prepReq) //TODO[sharon]: do error checking; Note: should return new value?

	// If majority is not reached, sleep for a while and try again
	// TODO: check whether should retry must return an error if no connection or something
	pn.ShouldRetry(numAccepted, value)

	accReq := pn.Proposer.CreateAcceptRequest(value)
	numAccepted, err = pn.DisseminateRequest(accReq)

	// If majority is not reached, sleep for a while and try again
	// TODO: check whether should retry must return an error if no connection or something
	pn.ShouldRetry(numAccepted, value)

	accReq.Type = CONSENSUS
	_, err = pn.DisseminateRequest(accReq)

	return success, err
}

// Sets up bidirectional RPC with all neighbours. Neighbours list is passed to the
// Paxos Node by the client.
// TODO: REVIEW PLEASE
func (pn *PaxosNode) BecomeNeighbours(ips []string) (err error) {
	pnAddr, err := net.ResolveTCPAddr("tcp", pn.Addr)
	if err != nil {
		fmt.Println("Error in resolving TCP address of PN")
		log.Fatal(err)
	}
	conn, err := net.ListenTCP("tcp", pnAddr)

	rpc.Register(pn)
	go rpc.Accept(conn)

	for _, ip := range ips {
		neighbourConn, err := rpc.Dial("tcp", ip)
		if err != nil {
			fmt.Println("Error in opening RPC connection with neighbour")
			log.Fatal(err)
		}
		connected := false
		//neighbourConn.Call("PaxosNode.AcceptNeighbourConnection", pnAddr, &connected)
		neighbourConn.Call("PaxosNodeInstance.ConnectRemoteNeighbour", pnAddr, &connected)

		// Add ip to connectedNbrs and add the connection to Neighbours map
		// after bidirectional RPC connection establishment is successful
		if connected {
			pn.NbrAddrs = append(pn.NbrAddrs, ip)
			pn.Neighbours[ip] = neighbourConn
		}
	}
	return err
}

// Disseminates a message to all neighbours. This includes prepare and accept requests.
//TODO[sharon]: Figure out best name for number field and add as param. Might be RPC
func (pn *PaxosNode) DisseminateRequest(prepReq Message) (numAccepted int, err error) {
	numAccepted = 0
	respReq := prepReq
	switch prepReq.Type {
	case PREPARE:
		for k, v := range pn.Neighbours {
			e := v.Call("PaxosNodeInstance.ProcessPrepareRequest", prepReq, &respReq)
			if e != nil {
				delete(pn.Neighbours, k)
				pn.RemoveNbrAddr(k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				if prepReq.Equals(&respReq) {
					numAccepted++
				}
			}
		}
	case ACCEPT:
		for k, v := range pn.Neighbours {
			e := v.Call("PaxosNodeInstance.ProcessAcceptRequest", prepReq, &respReq)
			if e != nil {
				delete(pn.Neighbours, k)
				pn.RemoveNbrAddr(k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				if prepReq.Equals(&respReq) {
					numAccepted++
				}
			}
		}
	case CONSENSUS:
		for k, v := range pn.Neighbours {
			e := v.Call("PaxosNodeInstance.ProcessLearnRequest", prepReq, &respReq)
			if e != nil {
				delete(pn.Neighbours, k)
				pn.RemoveNbrAddr(k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				if prepReq.Equals(&respReq) {
					numAccepted++
				}
			}
		}
	}

	return numAccepted, err
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
func (pn *PaxosNode) CountForDecisions (m * Message) {
	numSeen := pn.Learner.Decisions (m)
	if pn.IsMajority(numSeen) {
		pn.Learner.LearnValue(m) // this should write to the log TODO: expansion make learner return next round
	}
}

// This method sets up the bi-directional RPC. A new PN joins the network and will
// establish an RPC connection with each of the other PNs
func (pn *PaxosNode) AcceptNeighbourConnection(addr string, result *bool) (err error) {
	neighbourConn, err := rpc.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error in opening RPC connection with a new neighbour that connected to PN")
		log.Fatal(err)
	}
	pn.NbrAddrs = append(pn.NbrAddrs, addr)
	pn.Neighbours[addr] = neighbourConn
	*result = true
	return nil
}

func (pn *PaxosNode) ShouldRetry(numAccepted int, value string) {
	if !pn.IsMajority(numAccepted) {
		time.Sleep(SLEEPTIME)
		pn.WriteToPaxosNode(value)
	}
}

func (pn *PaxosNode) RemoveNbrAddr(ip string) {
	for i, v := range pn.NbrAddrs {
		if v == ip {
			pn.NbrAddrs = append(pn.NbrAddrs[:i], pn.NbrAddrs[i+1:]...)
			break
		}
	}
}

