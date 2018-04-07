package learner

import (
	"consensuslib/errors"
	"consensuslib/message"
	"filelogger/singletonlogger"
	"fmt"
	"paxostracker"
)

type Message = message.Message

type MessageAccepted struct {
	M     *Message
	Times int
}

type LearnerRole struct {
	Accepted     *SyncLog
	Log          []Message
	CurrentRound int // Should start at 0
}

type LearnerInterface interface {
	/**
	 * This is the interface that the PaxosNode uses to talk to the Learner.
	 **/

	// This method is used to set the initial log state when a PN joins
	// the network and learns of the majority log state from other PNs
	InitializeLog(log []Message) (err error)

	// Get this learner's current version of the PN log
	GetCurrentLog() (log []Message, err error)

	// Get the number of times this particular message ID has been accepted by this Learner
	NumAlreadyAccepted(m *Message) int

	// Writes the given message to the Log at the given current round index to log,
	// and auto-increments the log index. Returns the new CurrentRound index.
	LearnValue(m *Message) (currentRoundIndex int, err error)
}

func NewLearner() LearnerRole {
	syncLog := NewSyncLog()
	learner := LearnerRole{Accepted: syncLog, Log: make([]Message, 0), CurrentRound: 0}
	return learner
}

func (l *LearnerRole) InitializeLog(log []Message) (err error) {
	singletonlogger.Debug(fmt.Sprintf("[learner] Initializing log with size %v", len(log)))
	l.Log = log
	l.CurrentRound = len(log)
	singletonlogger.Debug(fmt.Sprintf("[learner] Initializing next round %v", l.CurrentRound))
	return nil
}

func (l *LearnerRole) GetCurrentLog() ([]Message, error) {
	return l.Log, nil
}

func (l *LearnerRole) NumAlreadyAccepted(m *Message) int {
	if accepted, ok := l.Accepted.Load(m.ID); ok {
		accepted.Times++
		return accepted.Times
	} else {
		l.Accepted.Store(m.ID, &MessageAccepted{m, 1})
		return 1
	}
}

func (l *LearnerRole) LearnValue(m *Message) (currentRoundIndex int, err error) {
	paxostracker.Learn(uint64(currentRoundIndex))
	singletonlogger.Debug(fmt.Sprintf("[learner] Writing value'%v'to round %v", m.Value, l.CurrentRound))
	if len(l.Log) > l.CurrentRound {
		// Since Learner manages this state, this should theoretically never happen...
		return l.CurrentRound, errors.ValueForRoundInLogExistsError(l.CurrentRound)
	} else {
		if l.inLog(m) {
			return m.RoundNum + 1, nil
		}
		l.Log = append(l.Log, *m)
		singletonlogger.Debug(fmt.Sprintf("[learner] Wrote value %v to log at index %v", l.Log[l.CurrentRound], l.CurrentRound))
		paxostracker.Idle(l.Log[l.CurrentRound].Value)
		l.CurrentRound++
		newInd := m.RoundNum + 1
		return newInd, nil
	}
}

func (l *LearnerRole) inLog(m *Message) bool {
	for _, v := range l.Log {
		if v.MsgHash == m.MsgHash {
			return true
		}
	}
	return false
}
