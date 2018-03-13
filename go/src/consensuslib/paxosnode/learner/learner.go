package learner

type Learner struct {
	Log []ConsensusValue
}

type LearnerInterface interface {
	GetCurrentLog() (log []string, err error)
	LearnConsensusValue() (learned bool, err error)
}