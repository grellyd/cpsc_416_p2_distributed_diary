package learner

import (
	"consensuslib"
)

type Message = consensuslib.Message

type LearnerRole struct {
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
