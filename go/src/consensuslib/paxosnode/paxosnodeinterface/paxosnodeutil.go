package paxosnodeinterface

import (
	. "consensuslib"
	"net/rpc"
	"net"
	"fmt"
	"log"
)

//type PaxosNodeInstance int

// Handles the entire process of proposing a value and trying to achieve consensus
//TODO[sharon]: update parameters as needed.
func (pn *PaxosNode) WriteToPaxosNode(value string) (success bool, err error) {
	prepReq := pn.Proposer.CreatePrepareRequest()
	numAccepted, err := DisseminateRequest(prepReq, pn.Neighbours) //TODO[sharon]: do error checking


	if !pn.IsMajority(numAccepted) {
		// TODO[sharon]: Handle not-majority. Quit or retry?
	}

	accReq := pn.Proposer.CreateAcceptRequest(value)
	numAccepted, err = DisseminateRequest(accReq, pn.Neighbours)

	if !pn.IsMajority(numAccepted) {
		// TODO[sharon]: Handle not-majority. Quit or retry?
	}

	accReq.Type = CONSENSUS
	_, err = DisseminateRequest(accReq, pn.Neighbours)

	return success, err
}

// Sets up bidirectional RPC with all neighbours. Neighbours list is passed to the
// Paxos Node by the client.
// TODO: REVIEW PLEASE
func (pn *PaxosNode) BecomeNeighbours(ips []string) (err error) {
	pnAddr, err := net.ResolveTCPAddr("tcp", pn.Addr)
	if (err != nil) {
		fmt.Println("Error in resolving TCP address of PN")
		log.Fatal(err)
	}
	conn, err := net.ListenTCP("tcp", pnAddr)

	rpc.Register(pn)
	go rpc.Accept(conn)

	for _, ip := range ips {
		neighbourConn, err := rpc.Dial("tcp", ip)
		if (err != nil) {
			fmt.Println("Error in opening RPC connection with neighbour")
			log.Fatal(err)
		}
		connected := false
		neighbourConn.Call("PaxosNode.AcceptNeighbourConnection", pnAddr, &connected)

		// Add ip to connectedNbrs and add the connection to Neighbours map
		// after bidirectional RPC connection establishment is successful
		if connected {
			pn.NbrAddrs = append(pn.NbrAddrs, ip)
			pn.Neighbours[ip] = neighbourConn
		}
	}
	return nil
}

// Disseminates a message to all neighbours. This includes prepare and accept requests.
//TODO[sharon]: Figure out best name for number field and add as param. Might be RPC
func DisseminateRequest(prepReq Message, neighbours map[string]*rpc.Client) (numAccepted int, err error) {
	numAccepted = 0
	switch prepReq.Type {
	case PREPARE :
		for k,v := range neighbours {
			e := v.Call("PaxosNodeInstance.ProcessPrepareRequest", prepReq, &prepReq)
			if e != nil {
				 delete(neighbours, k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				numAccepted++
			}
		}
	case ACCEPT :
		for k,v := range neighbours {
			e := v.Call("PaxosNodeInstance.ProcessAcceptRequest", prepReq, &prepReq)
			if e != nil {
				delete(neighbours, k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				numAccepted++
			}
		}
	case CONSENSUS :
		for k,v := range neighbours {
			e := v.Call("PaxosNodeInstance.ProcessLearnRequest", prepReq, &prepReq)
			if e != nil {
				delete(neighbours, k)
			} else {
				// TODO: check on what prepare request it returned, maybe to implement additional response OK/NOK
				// for now just a stub which increases count anyway
				numAccepted++
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
	if n > len(pn.Neighbours) / 2 {
		return true
	}
	return false
}

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
