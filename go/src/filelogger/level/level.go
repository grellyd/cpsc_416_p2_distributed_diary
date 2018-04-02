package level

// Level of a log statement
type Level string

// Const strings are normalised to the same length
const (
	// DEBUG - extra debug information
	DEBUG Level = "Debug  "
	// INFO - informational
	INFO Level = "Info   "
	// WARNING - indicator of something going wrong
	WARNING Level = "Warning"
	// ERROR - something has gone wrong, but the application can continue
	ERROR Level = "Error  "
	// FATAL - something has gone wrong, and the application cannot continue
	FATAL Level = "Fatal  "
)
