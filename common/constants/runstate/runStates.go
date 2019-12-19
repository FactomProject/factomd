package runstate

import "fmt"

type RunState int

const (
	New      RunState = iota // State of a newly created Factomd object
	Booting  RunState = iota // State when starting up the server
	Running  RunState = iota // State when doing processing
	Stopping RunState = iota // State when shutdown has been called but not finished
	Stopped  RunState = iota // State when shutdown has been completed
)

// IsTerminating returns true if factomd is terminated or in the process of terminating
func (runState RunState) IsTerminating() bool {
	return runState >= Stopping
}

// String returns the current run state as a string
func (runState RunState) String() string {
	switch runState {
	case New:
		return "New"
	case Booting:
		return "Booting"
	case Running:
		return "Running"
	case Stopping:
		return "Stopping"
	case Stopped:
		return "Stopped"
	default:
		return fmt.Sprintf("Unknown state %d", int(runState))
	}
}
