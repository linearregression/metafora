package statemachine

import (
	"errors"
	"time"
)

// ExceededErrorRate is returned by error handlers in an Error Message when
// retry logic has been exhausted for a handler and it should transition to
// Failed.
var ExceededErrorRate = errors.New("exceeded error rate")

// Err represents an error that occurred while a stateful handler was running.
type Err struct {
	Time time.Time `json:"timestamp"`
	Err  string    `json:"error"`
}

// ErrHandler functions should return Run, Sleep, or Fail messages depending on
// the rate of errors.
//
// Either ErrHandler and/or StateStore should trim the error slice to keep it
// from growing without bound.
type ErrHandler func(taskID string, errs []Err) (Message, []Err)

const (
	DefaultErrLifetime = -4 * time.Hour
	DefaultErrMax      = 8
)

// DefaultErrHandler returns a Fail message if 8 errors have occurred in 4
// hours. Otherwise it enters the Sleep state for 10 minutes before trying
// again.
func DefaultErrHandler(_ string, errs []Err) (Message, []Err) {
	recent := time.Now().Add(DefaultErrLifetime)
	strikes := 0
	for _, err := range errs {
		if err.Time.After(recent) {
			strikes++
		}
	}

	if strikes >= DefaultErrMax {
		// Return a new error to transition to Failed as well as the original
		// errors to store what caused this failure.
		return Message{Code: Error, Err: ExceededErrorRate}, errs
	}
	keeperrs := errs
	if len(keeperrs) > DefaultErrMax {
		keeperrs = keeperrs[len(keeperrs)-DefaultErrMax:]
	}
	until := time.Now().Add(10 * time.Minute)
	return Message{Code: Sleep, Until: &until}, keeperrs
}
