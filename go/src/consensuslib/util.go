package consensuslib

import "time"

type MsgType int

// generates a new message
type Message struct {
	ID             uint64  // unique ID for the paxos NW
	Type           MsgType // msgType should only be 'prepare' or 'accept'. 'prepare' messages should have empty value field
	Value          string  // value that needs to be written into log
	FromProposerID string  // Proposer's ID to distinguish when same ID message arrived
}

func NewMessage(id uint64, msgType MsgType, val string, pid string) Message {
	m := Message{
		id,
		msgType,
		val,
		pid,
	}
	return m
}

const (
	PREPARE MsgType = iota
	ACCEPT
	CONSENSUS
)

const SLEEPTIME = 100 * time.Millisecond
