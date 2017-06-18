/*
Package circuitbreaker implements a 3-state proxy to improve resilience of the application and its dependencies
See docs: https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker
*/
package circuitbreaker

import (
	"fmt"
	"time"
)

// defaults
var (
	defaultOpenStatusPeriod   = 1 * time.Minute
	defaultSuccessThreshold   = uint32(5) // number of consecutive success to transition from half-open to closed state
	defaultFailureThreshold   = uint32(5) // number of consecutive failures to transition from closed to open state
	defaultCounterResetPeriod = 1 * time.Minute

	ErrRequestDisabled = fmt.Errorf("requests are temporarily disabled by the circuit breaker")
)

type (
	// RequestFunc type for the request executor
	RequestFunc func() (interface{}, error)
	// Option allows to extend default circuit breaker
	Option func(*CircuitBreaker)

	// Counter statistics counter
	// TODO: set correct public/private
	Counter struct {
		failure     uint32
		success     uint32
		lastFail    time.Time
		lastSuccess time.Time
		resetPeriod time.Duration
	}
)

// Fail increases the consecutive failure counter
// if the last failure increase happened long ago,
// counter for failure should be set to 1
// as a correctness measure for rarely used services
func (c *Counter) Fail() uint32 {
	if time.Since(c.lastFail) > c.resetPeriod {
		c.failure = 0
	}
	c.lastFail = time.Now()
	c.failure++
	return c.failure
}

// Success increases the consecutive success counter
// if the last success increase happened long ago,
// counter for success should be set to 1
// as a correctness measure for rarely used services
func (c *Counter) Success() uint32 {
	if time.Since(c.lastSuccess) > c.resetPeriod {
		c.success = 0
	}
	c.lastSuccess = time.Now()
	c.success++
	return c.success
}

// Reset resets all stats
func (c *Counter) Reset() {
	c.lastFail = time.Now().Add(-24 * time.Hour)
	c.lastSuccess = time.Now().Add(-24 * time.Hour)
	c.failure = 0
	c.success = 0
}

// CircuitBreaker implements circuit breaker :D
type CircuitBreaker struct {
	counter          *Counter
	state            *State
	logger           Logger
	failureThreshold uint32
	successThreshold uint32
}

// New initializes default circuit breaker
func New(settings ...Option) *CircuitBreaker {
	cb := &CircuitBreaker{
		counter:          &Counter{},
		state:            NewState(),
		logger:           NoopLogger{},
		failureThreshold: defaultFailureThreshold,
		successThreshold: defaultSuccessThreshold,
	}

	for _, opt := range settings {
		opt(cb)
	}

	return cb
}

// WithFailureThreshold overwrites default value for number of failed request required before
// circuit breaker will enter open state
func WithFailureThreshold(t uint32) Option {
	return func(cb *CircuitBreaker) {
		cb.failureThreshold = t
	}
}

// WithSuccessThreshold overwrites default value for number of successfuly request required before
// circuit breaker will enter Closed state from Half Open state
func WithSuccessThreshold(t uint32) Option {
	return func(cb *CircuitBreaker) {
		cb.successThreshold = t
	}
}

// WithCounterResetPeriod defines the period after which counter will reset its failure/success counter
func WithCounterResetPeriod(t time.Duration) Option {
	return func(cb *CircuitBreaker) {
		cb.counter.resetPeriod = t
	}
}

// WithOpenPeriod defines the period for which circuit breaker can stay in open state
func WithOpenPeriod(t time.Duration) Option {
	return func(cb *CircuitBreaker) {
		cb.state.openPeriod = t
	}
}

// WithLogger allows to replace default no-op logger
func WithLogger(l Logger) Option {
	return func(cb *CircuitBreaker) {
		cb.logger = l
	}
}

// Exec is the wrapper for the request which encapsulates the circuit breaker logic
func (cb *CircuitBreaker) Exec(fn RequestFunc) (interface{}, error) {
	status := cb.state.Status()
	cb.logger.Debug("current status: ", status)
	switch status {
	case StatusClosed:
		res, err := fn()
		if err != nil {
			cb.logger.Error("request failed with ", err)
			if cb.counter.Fail() > cb.failureThreshold {
				cb.logger.Info("circuitbreaker entering open state, no request can be executed for the next ", cb.state.openPeriod)
				cb.state.Set(StatusOpen)
			}
			return nil, err
		}
		cb.logger.Debug("request succeeded")
		cb.counter.Success()
		return res, nil
	case StatusHalfOpen:
		// half open is intermediate state, where any failure will set back the circuitbreaker into open state
		// if required number of success responses are received circuitbreaker goes back to the closed state
		res, err := fn()
		if err != nil {
			cb.counter.Fail()
			//forgot to update the state
			cb.state.Set(StatusOpen)
			return nil, err
		}
		if cb.counter.Success() > cb.successThreshold {
			cb.state.Set(StatusClosed)
		}
		return res, nil
	case StatusOpen:
		return nil, ErrRequestDisabled
	}
	return nil, nil
}

// Reset allows to reset the state to the default
// needed for manual override
func (cb *CircuitBreaker) Reset() {
	cb.counter.Reset()
	cb.state.Reset()
}
