package message

import "time"

const (
	PREPARE MsgType = iota
	ACCEPT
	CONSENSUS
)

const SLEEPTIME = 100 * time.Millisecond

type MsgType int

// generates a new message
type Message struct {
	ID             uint64  // unique ID for the paxos NW
	Type           MsgType // msgType should only be 'prepare' or 'accept'. 'prepare' messages should have empty value field
	Value          string  // value that needs to be written into log
	FromProposerID string  // Proposer's ID to distinguish when same ID message arrived
	RoundNum	   int	   // The number of the round the message is for
}

func NewMessage(id uint64, msgType MsgType, val string, pid string, roundNum int) Message {
	m := Message{
		id,
		msgType,
		val,
		pid,
		roundNum,
	}
	return m
}

func (m *Message) Equals(m1 *Message) bool {
	if m.ID == m1.ID && m.Value == m1.Value {
		return true
	}
	return false
}
