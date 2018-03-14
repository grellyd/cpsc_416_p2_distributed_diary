package learner

import (
	"consensuslib"
)

type Message = consensuslib.Message

type LearnerRole struct {

}

type Learner struct {
	Log []Message
}

type LearnerInterface interface {
	GetCurrentLog() (log []string, err error)
	LearnConsensusValue() (learned bool, err error)
}