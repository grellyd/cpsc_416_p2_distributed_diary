package paxostracker

import (
	"filelogger/singletonlogger"
	"paxostracker/state"
	"paxostracker/errors"
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

// NewPaxosTracker creates a new tracker
func NewPaxosTracker() (err error) {
	tracker = &PaxosTracker{
			currentState: state.Idle,
	}
	return nil
}

// Perpare request
func Perpare() error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	return tracker.transition(state.Preparing)
}

// Propose request
func Propose() error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	return tracker.transition(state.Proposing)
}

// Learn value
func Learn() error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	return tracker.transition(state.Learning)
}

// Idle return
func Idle() error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	return tracker.transition(state.Idle)
}


// AsTable returns the current state of the paxos process in easily consumable table form.
func (t *PaxosTracker) AsTable() string {
	return ""
}

// transition from the current state to another, and return error if not possible 
func (t *PaxosTracker) transition(to state.PaxosState) error {
	switch t.currentState {
	case state.Idle:
		if !to.OneOf([]state.PaxosState{state.Preparing, state.Promised}) {
			return errors.BadTransition("")
		}
		return nil
	case state.Preparing:
		if !to.OneOf([]state.PaxosState{state.Proposing}) {
			return errors.BadTransition("")
		}
		return nil
	case state.Proposing:
		if !to.OneOf([]state.PaxosState{state.Learning}) {
			return errors.BadTransition("")
		}
		return nil
	case state.Learning:
		if !to.OneOf([]state.PaxosState{state.Idle}) {
			return errors.BadTransition("")
		}
		return nil
	case state.Promised:
		if !to.OneOf([]state.PaxosState{state.Accepted}) {
			return errors.BadTransition("")
		}
		return nil
	case state.Accepted:
		if !to.OneOf([]state.PaxosState{state.Idle}) {
			return errors.BadTransition("")
		}
		return nil
	default:
		return errors.UnknownTransition("")
	}
}
