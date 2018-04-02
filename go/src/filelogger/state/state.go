package state

// State of the logger
type State int

const (
	// NORMAL - Print Info and above to console and disk
	NORMAL State = 0
	// QUIET - Print Nothing
	QUIET State = 1
	// NOWRITE - Normal Do not write to disk
	NOWRITE State = 2
	// DEBUGGING - Print all
	DEBUGGING State = 3
)
