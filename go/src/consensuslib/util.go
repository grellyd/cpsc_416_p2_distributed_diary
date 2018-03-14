package consensuslib

/* Replaced with the message

type ConsensusValue struct {
	Index int
	Value string
}

type Proposal struct {
	PrepareID int
	Value string
}*/

// generates a new message
type Message struct {
	ID uint64				// unique ID for the paxos NW
	MsgType string          // msgType should only be 'prepare' or 'accept'. 'prepare' messages should have empty value field
	Value string			// value that needs to be written into log
	FromProposerID string	// Proposer's ID to distinguish when same ID message arrived
}

func NewMessage (id uint64, msgType string, val string, pid string) Message {
	m := Message{
		id,
		msgType,
		val,
		pid,
	}
	return m
}
