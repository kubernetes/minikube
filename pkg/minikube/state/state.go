// Package state contains user-readable state information about a minikube cluster
package state

import "fmt"

type Condition int

const (
	Unknown Condition = iota

	// Nonexistent is when there is no evidence that this object exists
	Nonexistent

	// Unavailable is when the object should exist, but isn't available
	Unavailable

	// Warning is when the object is available, but there is a warning
	Warning

	// Error is when the object is available, but it probably does not work
	Error

	// Unauthorized is when we do not have access to check the object state
	Unauthorized

	// Starting is ...
	Starting

	// OK is when all evidence points to nominal operation
	OK

	// Pausing is when an object is being paused
	Pausing

	// Paused is ...
	Paused

	// Stopping is ...
	Stopping

	// Stopped is ...
	Stopped
)

func IsUnavailable(c Condition) bool {
	if c == Stopped || c == Stopping || c == Paused || c == Pausing || c == Unavailable {
		return true
	}
	return false
}

// Component is used for components, clusters, and nodes
type Component struct {
	Name           string
	Kind           string
	TransitionStep string

	Condition Condition
	Errors    []string
	Warnings  []string
	Messages  []string
}

// String returns a lazy string value for a component
func (c Component) String() string {
	return fmt.Sprintf("%+v", c)
}

// Merge merges two component states
func (c Component) Merge(b Component) Component {
	if b.Name != "" {
		c.Name = b.Name
	}
	if b.Kind != "" {
		c.Kind = b.Kind
	}

	if b.TransitionStep != "" {
		c.TransitionStep = b.TransitionStep
	}

	if b.Condition != Unknown {
		c.Condition = b.Condition
	}

	c.Errors = append(c.Errors, b.Errors...)
	c.Warnings = append(c.Warnings, b.Warnings...)
	c.Messages = append(c.Messages, b.Messages...)
	return c
}

// A collection
type Collection struct {
	Component

	Components map[string]Component
}

// LibMachineCondition converts a libmachine state to a condition
func LibMachineCondition(s string) Condition {
	switch s {
	case "none":
		return Nonexistent
	case "running":
		return OK
	}

	return Unknown
}
