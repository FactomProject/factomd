package runstate

import "fmt"

type RunState int

const (
	New      RunState = 0
	Booting  RunState = 1
	Running  RunState = 2
	Stopping RunState = 3
	Stopped  RunState = 4
)

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
