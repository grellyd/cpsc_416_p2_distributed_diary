package learner

import (
	"fmt"
	"consensuslib/message"
)

type Message = message.Message

type MessageAccepted struct {
	M *Message
	Times int

}

type LearnerRole struct {
	Accepted map[uint64] *MessageAccepted // variable for mapping the accepted messages to count
	Log []Message
}

// TODO: Is this struct necessary?
//type Learner struct {
//	Log []Message
//}

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
	l.Log = log
	return nil
}

func (l *LearnerRole) GetCurrentLog() ([]Message, error) {
	return l.Log, nil
}

func (l *LearnerRole) NumAlreadyAccepted(m *Message) int {
	fmt.Println("[Learner] in NumAlreadyAccepted")
	return 2
}

func (l *LearnerRole) LearnValue (m *Message)  {
	// stub
	fmt.Println("Learning value ", m.Value)
}

