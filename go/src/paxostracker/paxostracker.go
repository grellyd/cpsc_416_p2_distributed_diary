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

// global vars
var tracker *PaxosTracker
var completedRounds []PaxosRound
var currentRound *PaxosRound

// signal channels
var prepareBreak chan struct{}
var proposeBreak chan struct{}
var learnBreak chan struct{}
var idleBreak chan struct{}
var customBreak chan struct{}
var continuePaxos chan struct{}

// NewPaxosTracker creates a new tracker
func NewPaxosTracker() (err error) {
	tracker = &PaxosTracker{
			currentState: state.Idle,
	}
	prepareBreak = make(chan struct{})
	proposeBreak = make(chan struct{})
	learnBreak = make(chan struct{})
	idleBreak = make(chan struct{}) 
	customBreak = make(chan struct{})
	continuePaxos = make(chan struct{})
	return nil
}


// Prepare request
func Prepare(callerAddr string) error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}

	select {
	case <- prepareBreak:
		// blocks until continue channel is filled
		<- continuePaxos
	default:
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

	select {
	case <- proposeBreak:
		// blocks until continue channel is filled
		<- continuePaxos
	default:
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

	select {
	case <- learnBreak:
		// blocks until continue channel is filled
		<- continuePaxos
	default:
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
	
	select {
	case <- idleBreak:
		// blocks until continue channel is filled
		<- continuePaxos
	default:
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

// Custom pause point
func Custom() error {
	if tracker == nil {
		singletonlogger.Error("Error: PaxosTracker Uninitialised")
		return nil
	}
	select {
	case <- customBreak:
		<- continuePaxos
	default:
	}
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

// BreakNextPrepare will block on the next prepare call till continue
func BreakNextPrepare() error {
	prepareBreak <- struct{}{}
	return nil
}

// BreakNextPropose will block on the next propose call till continue
func BreakNextPropose() error {
	proposeBreak <- struct{}{}
	return nil
}

// BreakNextLearn will block on the next learn call till continue
func BreakNextLearn() error {
	learnBreak <- struct{}{}
	return nil
}

// BreakNextIdle will block on the next idle call till continue
func BreakNextIdle() error {
	idleBreak <- struct{}{}
	return nil
}

// BreakNextCustom will block on the next custom call till continue
func BreakNextCustom() error {
	customBreak <- struct{}{}
	return nil
}

// Continue the execution of paxos
func Continue() error {
	continuePaxos <- struct{}{}
	return nil
}

// AsTable returns the current state of the paxos process in human consumable table form.
func AsTable() string {
	rows := "| Initial Addr | AcceptedPrepare | AcceptedProposal | Value |\n"
	for _, round := range(completedRounds) {
		rows += round.AsRow()
	}
	var pstate state.PaxosState
	if tracker == nil {
		pstate = state.Idle
	} else {
		pstate = tracker.currentState
	}
	return fmt.Sprintf("\n======================\nCurrent State: %v\n======================\n%v", pstate, rows)
}
