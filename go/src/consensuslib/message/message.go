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
	MsgHash        string  // unique hash for the message
	Type           MsgType // msgType should only be 'prepare' or 'accept'. 'prepare' messages should have empty value field
	Value          string  // value that needs to be written into log
	FromProposerID string  // Proposer's ID to distinguish when same ID message arrived
	RoundNum       int     // The number of the round the message is for
	Bounces        int     // TTL for the message
}

// generates a new message
func NewMessage(id uint64, msgHash string, msgType MsgType, val string, pid string, roundNum, ttl int) Message {
	m := Message{
		id,
		msgHash,
		msgType,
		val,
		pid,
		roundNum,
		ttl,
	}
	return m
}

// checks whether or not messages are equal based on the unique hash
func (m *Message) Equals(m1 *Message) bool {
	if m.MsgHash == m1.MsgHash {
		return true
	}
	return false
}
