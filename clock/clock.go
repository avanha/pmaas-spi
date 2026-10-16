// Package clock is the server's clock: an entity that, once started, broadcasts an event each time the
// wall-clock minute, hour, day, month or year changes, so that something can be done at a time of day without
// every plugin running a timer of its own.
package clock

import (
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi/events"
)

// Clock is the clock entity. It's registered by the server itself, always, but it doesn't run until it's started.
type Clock interface {
	// Start has the clock begin broadcasting events. Nothing is broadcast for the minute it's started in, only
	// for the changes after it. Starting a clock that is already running does nothing.
	Start() error

	// Stop has the clock stop broadcasting events. It returns once the clock has finished, so that no event is
	// broadcast after it. Stopping a clock that isn't running does nothing.
	Stop() error

	// Running is whether the clock is started.
	Running() (bool, error)
}

var ClockType = reflect.TypeFor[Clock]()

// ClockEntityStubAvailableEvent is how the assembly is handed the clock's stub, once the clock has been
// registered. It's delivered on the server's goroutine, the one that called Run.
type ClockEntityStubAvailableEvent struct {
	events.EntityEvent
	EntityStub Clock
}

// The events below are broadcast by a running clock, in the local time zone, each one carrying the minute that
// has just begun: its seconds and below are zero. When more than one thing changes at the same moment, say the
// last minute of the year, the events are broadcast from the largest to the smallest: the year, the month, the
// day, the hour and then the minute. A receiver of the minute is therefore never told about a new minute before
// it has been told about the new hour it's in.
//
// A clock that is kept from running for a while, because the machine was asleep or the time was changed, doesn't
// replay the minutes it missed: it broadcasts once for the time it finds, and what changed since it last did.
// An hour that was slept through is one hour change.

// MinuteChangedEvent is broadcast every minute.
type MinuteChangedEvent struct {
	events.EntityEvent
	Time time.Time
}

// HourChangedEvent is broadcast when a new hour begins.
type HourChangedEvent struct {
	events.EntityEvent
	Time time.Time
}

// DayChangedEvent is broadcast when a new day begins.
type DayChangedEvent struct {
	events.EntityEvent
	Time time.Time
}

// MonthChangedEvent is broadcast when a new month begins.
type MonthChangedEvent struct {
	events.EntityEvent
	Time time.Time
}

// YearChangedEvent is broadcast when a new year begins.
type YearChangedEvent struct {
	events.EntityEvent
	Time time.Time
}

// IsClockEvent is an events.EventPredicate for the events above, all of them. A receiver that wants only some of
// them tests for the types it wants itself.
func IsClockEvent(eventInfo *events.EventInfo) bool {
	switch eventInfo.Event.(type) {
	case MinuteChangedEvent, HourChangedEvent, DayChangedEvent, MonthChangedEvent, YearChangedEvent:
		return true
	default:
		return false
	}
}
