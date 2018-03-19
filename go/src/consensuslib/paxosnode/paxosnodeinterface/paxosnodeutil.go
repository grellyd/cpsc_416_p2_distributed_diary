package paxosnodeinterface

import (
	. "consensuslib"
	"fmt"
	//"log"
	//"net"
	"net/rpc"
	"time"
)

//type PaxosNodeInstance int

// Handles the entire process of proposing a value and trying to achieve consensus
//TODO[sharon]: update parameters as needed.
func (pn *PaxosNode) WriteToPaxosNode(value string) (success bool, err error) {
	fmt.Println("[paxosnodeutil] Writing to paxos")
	prepReq := pn.Proposer.CreatePrepareRequest()
	fmt.Printf("[paxosnodeutil] Prepare request is id: %d , val: %s, type: %d \n", prepReq.ID, prepReq.Value, prepReq.Type)
	numAccepted, err := pn.DisseminateRequest(prepReq)
	fmt.Println("[paxosnodeutil] Pledged to accept ", numAccepted)
	// TODO: Unsure if err from DisseminateRequest should bubble up to client. Previous Note: should return new value?
	if err != nil {
		return false, err
	}

	// If majority is not reached, sleep for a while and try again
	// TODO: check whether should retry must return an error if no connection or something
	pn.ShouldRetry(numAccepted, value)

	accReq := pn.Proposer.CreateAcceptRequest(value)
	fmt.Printf("[paxosnodeutil] Accept request is id: %d , val: %s, type: %d \n", accReq.ID, accReq.Value, accReq.Type)
	numAccepted, err = pn.DisseminateRequest(accReq)
	if err != nil {
		return false, err
	}
	fmt.Println("[paxosnodeutil] Accepted ", numAccepted)
	// If majority is not reached, sleep for a while and try again
	// TODO: check whether should retry must return an error if no connection or something
	pn.ShouldRetry(numAccepted, value)

	accReq.Type = CONSENSUS
	_, err = pn.DisseminateRequest(accReq)
	if err != nil {
		return false, err
	}

	return success, nil
}

// Sets up bidirectional RPC with all neighbours. Neighbours list is passed to the
// Paxos Node by the client.
func (pn *PaxosNode) BecomeNeighbours(ips []string) (err error) {
	// Commented out since we already establish RPC listener at client.go
	// Otherwise it will create different IP:Port combination unknown to the Server
	/*pnAddr, err := net.ResolveTCPAddr("tcp", pn.Addr)
	if err != nil {
		fmt.Println("Error in resolving TCP address of PN")
		log.Fatal(err)
	}
	conn, err := net.ListenTCP("tcp", pnAddr)

	rpc.Register(pn)
	go rpc.Accept(conn)*/

	for _, ip := range ips {
		neighbourConn, err := rpc.Dial("tcp", ip)
		if err != nil {
			return NeighbourConnectionError(ip)
		}
		connected := false
		//err = neighbourConn.Call("PaxosNodeInstance.ConnectRemoteNeighbour", pnAddr, &connected)
		err = neighbourConn.Call("PaxosNodeInstance.ConnectRemoteNeighbour", pn.Addr, &connected)

		// Add ip to connectedNbrs and add the connection to Neighbours map
		// after bidirectional RPC connection establishment is successful
		if connected {
			fmt.Println("[paxosnodeutil]: connected to the nbr")
			pn.NbrAddrs = append(pn.NbrAddrs, ip)
			if pn.Neighbours == nil {
				pn.Neighbours = make(map[string]*rpc.Client,0)
			}
			pn.Neighbours[ip] = neighbourConn
		}
	}
	return nil
}


// This method sets up the bi-directional RPC. A new PN joins the network and will
// establish an RPC connection with each of the other PNs
func (pn *PaxosNode) AcceptNeighbourConnection(addr string, result *bool) (err error) {
	neighbourConn, err := rpc.Dial("tcp", addr)
	if err != nil {
		return NeighbourConnectionError(addr)
	}
	pn.NbrAddrs = append(pn.NbrAddrs, addr)
	if pn.Neighbours == nil {
		pn.Neighbours = make(map[string]*rpc.Client,0)
	}
	pn.Neighbours[addr] = neighbourConn
	*result = true
	return nil
}

func (pn *PaxosNode) RemoveNbrAddr(ip string) {
	for i, v := range pn.NbrAddrs {
		if v == ip {
			pn.NbrAddrs = append(pn.NbrAddrs[:i], pn.NbrAddrs[i+1:]...)
			break
		}
	}
}

// Disseminates a message to all neighbours. This includes prepare and accept requests.
//TODO[sharon]: Figure out best name for number field and add as param. Might be RPC
func (pn *PaxosNode) DisseminateRequest(prepReq Message) (numAccepted int, err error) {
	fmt.Println("[paxosnodeutil] Disseminate request")
	numAccepted = 0
	respReq := prepReq
	switch prepReq.Type {
	case PREPARE:
		fmt.Println("[paxosnodeutil] PREPARE")
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
		// last send it to ourselves
		pn.Acceptor.ProcessPrepare(prepReq)
		if prepReq.Equals(&respReq) {
			numAccepted++
		}
	case ACCEPT:
		fmt.Println("[paxosnodeutil] ACCEPT")
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
		// last send it to ourselves
		pn.Acceptor.ProcessAccept(prepReq)
		if prepReq.Equals(&respReq) {
			numAccepted++
			fmt.Println("[paxosnodeutil] saying accepted for myself")
			go pn.SayAccepted(&prepReq)
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
	default:
		return -1, InvalidMessageTypeError(prepReq)
	}

	return numAccepted, err
}

// Notifies all learners that request was accepted
func (pn *PaxosNode) SayAccepted (m *Message) {
	var counted bool
	for k, v := range pn.Neighbours {
		e := v.Call("PaxosNodeInstance.NotifyAboutAccepted", m, &counted)
		if e != nil {
			delete(pn.Neighbours, k)
			pn.RemoveNbrAddr(k)
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
func (pn *PaxosNode) CountForNumAlreadyAccepted(m * Message) {
	fmt.Println("[paxosnodeutil] in CountForNumAlreadyAccepted")
	numSeen := pn.Learner.NumAlreadyAccepted(m)
	if pn.IsMajority(numSeen) {
		pn.Learner.LearnValue(m) // this should write to the log TODO: expansion make learner return next round
	}
}

func (pn *PaxosNode) ShouldRetry(numAccepted int, value string) {
	if !pn.IsMajority(numAccepted) {
		time.Sleep(SLEEPTIME)
		pn.WriteToPaxosNode(value)
	}
}