package learner

import (
	"consensuslib"
)

type Message = consensuslib.Message
type MessageAccepted struct {
	M *Message
	Times int

}

type LearnerRole struct {
	//Accepted map[uint64] *MessageAccepted // variable for mapping the accepted messages to count
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
	/*if val, ok := l.Accepted[m.ID]; ok {
		val.Times++
		return val.Times
	}
	ma := MessageAccepted{m, 1}
	l.Accepted[m.ID] = &ma
	return 1*/
	return 1
}

func (l *LearnerRole) LearnValue (m *Message)  {
	// stub
}

