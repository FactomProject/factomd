package worker

// a single execution - return false to signal parent loop to break
type step func() bool

// wrap the execution of sequential steps
func RunSteps(steps ...step) bool {
	for i := range steps {
		if !steps[i]() {
			return false
		}
	}
	return true
}
