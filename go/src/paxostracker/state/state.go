package state

// PaxosState represents a state in paxos
type PaxosState string

const (
	// Idle is not in a round
	Idle PaxosState = "Idle"

	// Active States
	// 	Happy Path: Idle -> Preparing -> Proposing -> Learning -> Idle

	// Preparing is while disseminating out the prepare request
	Preparing PaxosState = "Preparing"
	// Proposing is while disseminating out the accept request
	Proposing PaxosState = "Proposing"
	// Learning is while saving the value to disk
	Learning PaxosState = "Learning"

	// Passive States
	// 	Happy Path: Idle -> Promised -> Accepted -> Idle

	// Promised is after receiving and approving a prepare request
	Promised PaxosState = "Promised"
	// Accepted is after receiving and approving a propose request
	Accepted PaxosState = "Accepted"

	// Error is when something has gone wrong. Only goes to Idle
	Error PaxosState = "Error"
)

// OneOf determines if the given state is contained with in the collection
func (s *PaxosState) OneOf(states []PaxosState) bool {
	for _, state := range states {
		if state == *s {
			return true
		}
	}
	return false
}
