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
var preparePause chan struct{}
var proposePause chan struct{}
var learnPause chan struct{}
var idlePause chan struct{}
var customPause chan struct{}
var continuePaxos chan struct{}

// NewPaxosTracker creates a new tracker
func NewPaxosTracker() (err error) {
	tracker = &PaxosTracker{
			currentState: state.Idle,
	}
	preparePause = make(chan struct{})
	proposePause = make(chan struct{})
	learnPause = make(chan struct{})
	idlePause = make(chan struct{}) 
	customPause = make(chan struct{})
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
	case <- preparePause:
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
	case <- proposePause:
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
	case <- learnPause:
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
	case <- idlePause:
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
	case <- customPause:
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

// PauseNextPrepare will block on the next prepare call till continue
func PauseNextPrepare() error {
	preparePause <- struct{}{}
	return nil
}

// PauseNextPropose will block on the next propose call till continue
func PauseNextPropose() error {
	proposePause <- struct{}{}
	return nil
}

// PauseNextLearn will block on the next learn call till continue
func PauseNextLearn() error {
	learnPause <- struct{}{}
	return nil
}

// PauseNextIdle will block on the next idle call till continue
func PauseNextIdle() error {
	idlePause <- struct{}{}
	return nil
}

// PauseNextCustom will block on the next custom call till continue
func PauseNextCustom() error {
	customPause <- struct{}{}
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
