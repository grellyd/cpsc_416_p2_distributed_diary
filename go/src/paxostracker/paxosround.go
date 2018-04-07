package paxostracker

import (
	"fmt"
)

// PaxosRound is a round of paxos
type PaxosRound struct {
	InitialAddr         string
	AcceptedPreparation uint64
	AcceptedProposal    uint64
	Value               string
	ErrorReason         string
}

// AsRow converts a round to a string row
func (r *PaxosRound) AsRow() string {
	if r.ErrorReason != "" {
		return fmt.Sprintf("| %s |", r.ErrorReason)
	}
	return fmt.Sprintf("| %s | %d | %d | %s |\n", r.InitialAddr, r.AcceptedPreparation, r.AcceptedProposal, r.Value)
}
