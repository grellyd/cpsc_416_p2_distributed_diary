package learner

import (
	"filelogger/singletonlogger"
	"consensuslib/errors"
	"consensuslib/message"
	"fmt"
)

type Message = message.Message

type MessageAccepted struct {
	M     *Message
	Times int
}

type LearnerRole struct {
	Accepted map[uint64]*MessageAccepted // variable for mapping the accepted messages to count
	Log      []Message
	CurrentRound int	// Should start at 0
}

type LearnerInterface interface {
	// This method is used to set the initial log state when a PN joins
	// the network and learns of the majority log state from other PNs
	InitializeLog(log []Message) (err error)

	GetCurrentLog() (log []Message, err error)

	LearnConsensusValue() (learned bool, err error)
}

func NewLearner() LearnerRole {
	learner := LearnerRole{ Accepted: make(map[uint64]*MessageAccepted, 0), Log: make([]Message, 0), CurrentRound: 0 }
	return learner
}

func (l *LearnerRole) InitializeLog(log []Message) (err error) {
	// TODO: What if the learner already has a filled in log? Does this suggest an error state?
	singletonlogger.Debug(fmt.Sprintf("[learner] Initializing log with size %v", len(log)))
	l.Log = log
	l.CurrentRound = len(log)
	singletonlogger.Debug(fmt.Sprintf("[learner] Initializing next round %v", l.CurrentRound))
	return nil
}

func (l *LearnerRole) GetCurrentLog() ([]Message, error) {
	return l.Log, nil
}

func (l *LearnerRole) GetCurrentRound() (int, error) {
	return l.CurrentRound, nil
}

func (l *LearnerRole) GetLogValue(round int) (string, error) {
	if len(l.Log) > round {
		return l.Log[round].Value, nil;
	} else {
		return "", errors.InvalidLogIndexError(round)
	}
}

func (l *LearnerRole) NumAlreadyAccepted(m *Message) int {
	if accepted, ok := l.Accepted[m.ID]; ok {
		accepted.Times++
	} else {
		l.Accepted[m.ID] = &MessageAccepted { m, 1 }
	}

	return l.Accepted[m.ID].Times
}

func (l *LearnerRole) LearnValue(m *Message) (newCurrentRoundIndex int, err error) {
	/*
		Writes the given message to the Log at the CurrentRound index to log,
		and auto-increments the log index. Returns the new CurrentRound.

		TODO: Do we want to auto-decrement in the learner, or should this be done elsewhere?
	 */
	singletonlogger.Debug(fmt.Sprintf("[learner] Writing value'%v'to round %v", m.Value, l.CurrentRound))
	if len(l.Log) > l.CurrentRound {
		// Since Learner manages this state, this should theoretically never happen...
        return l.CurrentRound, errors.ValueForRoundInLogExistsError(l.CurrentRound)
	} else {
		l.Log = append(l.Log, *m)
		singletonlogger.Debug(fmt.Sprintf("[learner] Wrote value %v to log at index %v", l.Log[l.CurrentRound], l.CurrentRound))
		l.CurrentRound++ // TODO: Once we have the concept of rounds
		return l.CurrentRound, nil
	}
}
