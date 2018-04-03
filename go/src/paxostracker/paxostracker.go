package paxostracker

import (
	"filelogger/singletonlogger"
	"paxostracker/state"
	"paxostracker/errors"
	"fmt"
)

/*
PaxosTracker is a global singleton instantiated per consensuslib client instance to track the state.
Paxostracker uses a DFA representation of the paxos process, and is activated by the consensuslib as it changes state. 
The paxostracker can output the current state at any time.
The paxostracker can add a wait before the next stage activation.
Each transition function call will return either nil or error.
*/

// PaxosTracker struct
type PaxosTracker struct {
	currentState state.PaxosState
}

var tracker *PaxosTracker
var completedRounds []PaxosRound
var currentRound *PaxosRound

// NewPaxosTracker creates a new tracker
func NewPaxosTracker() (err error) {
	tracker = &PaxosTracker{
			currentState: state.Idle,
	}
	return nil
}

// Prepare request
func Prepare(callerAddr string) error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	switch tracker.currentState {
	case state.Idle:
	default:
		return errors.BadTransition("")
	}
	currentRound = &PaxosRound{
		InitialAddr: callerAddr,
	}
	tracker.currentState = state.Preparing
	return nil
}

// Propose request
func Propose(acceptedPrep uint64) error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	switch tracker.currentState {
	case state.Preparing:
	default:
		return errors.BadTransition("")
	}
	currentRound.AcceptedPreparation = acceptedPrep
	tracker.currentState = state.Proposing
	return nil
}

// Learn value
func Learn(acceptedProp uint64) error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	switch tracker.currentState {
	case state.Proposing:
	default:
		return errors.BadTransition("")
	}
	currentRound.AcceptedProposal = acceptedProp
	tracker.currentState = state.Learning
	return nil
}

// Idle return
func Idle(finalValue string) error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	// check for valid transitions
	switch tracker.currentState {
	case state.Learning:
	case state.Accepted:
	default:
		return errors.BadTransition("")
	}
	currentRound.Value = finalValue
	tracker.currentState = state.Idle
	// save the completed round
	completedRounds = append(completedRounds, *currentRound)
	// reset current round
	currentRound = nil
	return nil
}

// Error transition
func Error(reason string) error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	// valid for all transitions
	currentRound.ErrorReason = reason
	tracker.currentState = state.Idle
	// save the completed round
	completedRounds = append(completedRounds, *currentRound)
	// reset current round
	currentRound = nil
	return nil
}



// AsTable returns the current state of the paxos process in human consumable table form.
func AsTable() string {
	rows := "| Initial Addr | AcceptedPrepare | AcceptedProposal | Value |\n"
	for _, round := range(completedRounds) {
		rows += round.AsRow()
	}
	return fmt.Sprintf("\n======================\nCurrent State: %v\n======================\n%v", tracker.currentState, rows)
}
