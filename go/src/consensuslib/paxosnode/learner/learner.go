package learner

import (
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
}

type LearnerInterface interface {
	// This method is used to set the initial log state when a PN joins
	// the network and learns of the majority log state from other PNs
	InitializeLog(log []Message) (err error)

	GetCurrentLog() (log []Message, err error)

	LearnConsensusValue() (learned bool, err error)
}

func NewLearner() LearnerRole {
	learner := LearnerRole{}
	return learner
}

func (l *LearnerRole) InitializeLog(log []Message) (err error) {
	// TODO: What if the learner already has a filled in log? Does this suggest an error state?
	l.Log = log
	return nil
}

func (l *LearnerRole) GetCurrentLog() ([]Message, error) {
	return l.Log, nil
}

func (l *LearnerRole) NumAlreadyAccepted(m *Message) int {
	if accepted, ok := l.Accepted[m.ID]; ok {
		accepted.Times++
	} else {
		l.Accepted[m.ID] = &MessageAccepted { m, 1 }
	}

	return l.Accepted[m.ID].Times
}

func (l *LearnerRole) LearnValue(m *Message) {
	// Initialize the log
	// Add to the log
	fmt.Println("Learning value ", m.Value)
}
