// Package automation is the contract between whatever runs automations (the automations plugin) and the
// automations themselves, which are plain code in an assembly or in a library of re-usable ones: turn the holiday
// lights on 15 minutes before sunset, and off at half past midnight.
//
// An automation says when it is due, by its Trigger, and what to do, by Run. It knows nothing about the server, and
// the server knows nothing about what it does: that is what lets the same automation be tested with fakes, and used
// by any assembly.
package automation

import (
	"errors"
	"fmt"
	"time"
)

// Trigger says when an automation is due, once a day. It is a pure function of the time: it knows nothing about what
// it is for, or whether it has been acted on. That is what lets whatever runs automations treat running on time,
// catching up after a pause and catching up after a restart as one question: when was this last due, and has it been
// done since?
//
// The times it is given and returns are in the same zone, which is the server's local time when the server asks about
// the time now. Each is to the minute, because that is how often the server checks.
type Trigger interface {
	// Previous is the latest time at or before t that the trigger was due. It reports false if it has not been due
	// for a long time, a year or so.
	Previous(t time.Time) (time.Time, bool)

	// Next is the earliest time after t that the trigger is due. It reports false if it is not due for a long time.
	Next(t time.Time) (time.Time, bool)

	// Description says when the trigger is due in words, for people: "at 17:00", "15 minutes before sunset".
	Description() string
}

// Automation is something to do at some time.
//
// Its methods are called from the server's own goroutine, one at a time, so an automation does not need to guard
// its state. That also means they must be quick: Run in particular should send its commands and return, as turning
// a switch on or off does, and not wait for anything to finish.
type Automation interface {
	// ID identifies the automation for as long as it exists. It is what the server remembers when the automation
	// last ran under, so it must not change when the automation is renamed, and must differ from every other
	// automation's in the same server.
	ID() string

	// Name is what the automation does, as a person would say it, for the page: "Turn on the holiday lights".
	Name() string

	// Description is more about it, for the page. It can be empty.
	Description() string

	// Trigger says when the automation is due.
	Trigger() Trigger

	// CatchUp is how long after it was due that it is still worth running. An automation that could not run when it
	// was due, because the server was not running or because Run failed, is run when the server next can, for this
	// long, and not after. Zero means it runs only in the minute it is due.
	CatchUp() time.Duration

	// Run does it. It returns a short note on what it did, for the page, which can be empty, or an error if it could
	// not, in which case it is run again each minute until it succeeds or its CatchUp has passed. It is called again
	// after it fails, and it can be called again after the server was stopped part way through it, so it should
	// leave things right however many times it is called: turn a switch on, don't toggle it.
	Run() (string, error)
}

// Validate is whether an automation can be run: the things the server needs of it. The automations plugin checks
// each it is given with it, and a constructor of automations can check with it that it is not making one that will
// be refused.
func Validate(a Automation) error {
	if a == nil {
		return errors.New("the automation is nil")
	}

	var problems []error

	if a.ID() == "" {
		problems = append(problems, errors.New("it has no ID"))
	}

	if a.Name() == "" {
		problems = append(problems, errors.New("it has no name"))
	}

	if a.Trigger() == nil {
		problems = append(problems, errors.New("it has no trigger"))
	}

	if a.CatchUp() < 0 {
		problems = append(problems, fmt.Errorf("its catch-up, %v, is negative", a.CatchUp()))
	}

	if len(problems) > 0 {
		return fmt.Errorf("automation %q is not valid: %w", a.ID(), errors.Join(problems...))
	}

	return nil
}
