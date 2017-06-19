package circuitbreaker

import (
	"sync"
	"time"
)

// Status is used by circuitbreaker to apply the correct behaviour for the request based:
// see docs for more info: https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker
type Status uint8

const (
	// StatusClosed - request receiver is expected to be in the positive status, there request is routed
	StatusClosed Status = iota
	// StatusOpen - request receiver failed to respond for multiple last requests, therefore we give it time to respawn
	StatusOpen
	// StatusHalfOpen - intermediate state, any failure will set back circuit breaker to open state
	StatusHalfOpen
)

// String to implement Stringer interface
// TODO: will be needed with injected logging
func (s Status) String() string {
	switch s {
	case StatusClosed:
		return "closed"
	case StatusOpen:
		return "open"
	case StatusHalfOpen:
		return "half-open"
	}
	return ""
}

// State indicates the state of the circuitbreaker
type State struct {
	lastOpen   time.Time
	openPeriod time.Duration
	status     Status
	sync.Mutex
}

// NewState initializes new State object
func NewState() *State {
	return &State{
		status: StatusClosed,
	}
}

// Status returns correct status for the given time
// it handles the special case when circuit breaker has set to open state
// far too long ago
func (s *State) Status() Status {
	s.Lock()
	defer s.Unlock()
	// check when state last entered open
	if s.status == StatusOpen {
		if time.Since(s.lastOpen) > s.openPeriod {
			s.status = StatusHalfOpen
		}
	}
	return s.status
}

// Set updates the status of the state
// additionally handles updating the flag indicating when circuitbreaker
// last entered open state
func (s *State) Set(status Status) {
	s.Lock()
	defer s.Unlock()
	// update lastOpen field
	if status == StatusOpen {
		s.lastOpen = time.Now()
	}

	s.status = status
}

// Reset resets all flags
func (s *State) Reset() {
	s.Lock()
	defer s.Unlock()

	s.status = StatusClosed
	s.lastOpen = time.Now().Add(-24 * time.Hour)
}
