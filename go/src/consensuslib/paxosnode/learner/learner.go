package learner

import (
	"consensuslib"
	"fmt"
)

type Message = consensuslib.Message
type MessageAccepted struct {
	M *Message
	Times int

}

type LearnerRole struct {
	Accepted map[uint64] *MessageAccepted // variable for mapping the accepted messages to count
}

func NewLearner() LearnerRole {
	learner := LearnerRole{}
	return learner
}

type Learner struct {
	Log []Message
}

type LearnerInterface interface {
	GetCurrentLog() (log []string, err error)
	LearnConsensusValue() (learned bool, err error)
}

func (l *LearnerRole) NumAlreadyAccepted(m *Message) int {
	fmt.Println("[Learner] in NumAlreadyAccepted")
	return 2
}

func (l *LearnerRole) LearnValue (m *Message)  {
	// stub
	fmt.Println("Learning value ", m.Value)
}

