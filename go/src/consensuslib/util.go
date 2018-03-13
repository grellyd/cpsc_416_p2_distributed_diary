package consensuslib

type ConsensusValue struct {
	Index int
	Value string
}

type Proposal struct {
	PrepareID int
	Value string
}
