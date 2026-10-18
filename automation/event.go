package automation

import (
	"errors"
	"fmt"

	"github.com/avanha/pmaas-spi/events"
)

// EventTrigger says which events make an event-driven automation due: a button being pressed, say. Where a Trigger is a
// function of the time, this is a function of an event, and the automation runs the moment one matches, not at the next
// minute.
type EventTrigger interface {
	// Matches is whether the event makes the automation due. It is asked of every event the server broadcasts, on the event
	// manager's own goroutine, so it has to be quick, never block, and look at nothing but the event and what was given to
	// the trigger to know the event by.
	Matches(info *events.EventInfo) bool

	// Description says what the trigger waits for in words, for people: "when the Couch button is pressed".
	Description() string
}

// EventAutomation is something to do when something happens, such as switching the lights on when a button is pressed. It
// is an Automation in all but when it is due: it has no time to be due at, so it has no catch-up and no schedule, and a run
// that failed is not tried again, since the moment it was wanted has gone.
//
// Run is called from the server's own goroutine, one at a time, like an Automation's, with the same advice: it should be
// quick, send its commands and return, and leave things right however often it is called.
type EventAutomation interface {
	// ID identifies the automation for as long as it exists. It must not change when the automation is renamed, and must
	// differ from every other automation's, event-driven or not, in the same server.
	ID() string

	// Name is what the automation does, as a person would say it, for the page.
	Name() string

	// Description is more about it, for the page. It can be empty.
	Description() string

	// EventTrigger says which events make it due.
	EventTrigger() EventTrigger

	// Run does it. It returns a short note on what it did, which can be empty, or an error if it could not. It may return
	// NotReady when the event was no reason to act, because a condition does not hold: that is recorded and not made
	// anything of.
	Run() (string, error)
}

// ValidateEvent is Validate for an EventAutomation.
func ValidateEvent(a EventAutomation) error {
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

	if a.EventTrigger() == nil {
		problems = append(problems, errors.New("it has no trigger"))
	}

	if len(problems) > 0 {
		return fmt.Errorf("automation %q is not valid: %w", a.ID(), errors.Join(problems...))
	}

	return nil
}
