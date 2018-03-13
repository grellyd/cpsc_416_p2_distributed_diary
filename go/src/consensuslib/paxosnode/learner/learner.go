package learner

type Learner struct {
	Log []Message
}

type LearnerInterface interface {
	GetCurrentLog() (log []string, err error)
	LearnConsensusValue() (learned bool, err error)
}