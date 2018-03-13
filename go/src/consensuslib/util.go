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
	Vale string				// value that needs to be written into log
	FromProposerID string	// Proposer's ID to distinguish when same ID message arrived
}

func NewMessage (id uint64, val string, pid string) Message {
	m := Message{
		id,
		val,
		pid,
	}
	return m
}
