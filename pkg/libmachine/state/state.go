package state

// State represents the state of a host
type State int

const (
	None State = iota
	Running
	Paused
	Saved
	Stopped
	Stopping
	Starting
	Error
	Timeout
)

var states = []string{
	"",
	"Running",
	"Paused",
	"Saved",
	"Stopped",
	"Stopping",
	"Starting",
	"Error",
	"Timeout",
}

// Given a State type, returns its string representation
func (s State) String() string {
	if int(s) >= 0 && int(s) < len(states) {
		return states[s]
	}
	return ""
}
