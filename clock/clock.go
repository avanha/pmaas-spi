// Package clock describes the server's clock: once the server is running it broadcasts an event each time the
// wall-clock minute, hour, day, month or year changes, so that something can be done at a time of day without
// every plugin running a timer of its own. Register an event receiver for them (see IsClockEvent); there is
// nothing to start.
//
// The events are broadcast in the local time zone, each one carrying the minute that has just begun: its seconds and
// below are zero. When more than one thing changes at the same moment, say the last minute of the year, the events
// are broadcast from the largest to the smallest: the year, the month, the day, the hour and then the minute. A
// receiver of the minute is therefore never told about a new minute before it has been told about the new hour it's
// in.
//
// A clock that is kept from running for a while, because the machine was asleep or the time was changed, doesn't
// replay the minutes it missed: it broadcasts once for the time it finds, and what changed since it last did. An
// hour that was slept through is one hour change.
package clock

import (
	"time"

	"github.com/avanha/pmaas-spi/events"
)

// MinuteChangedEvent is broadcast every minute.
type MinuteChangedEvent struct {
	Time time.Time
}

// HourChangedEvent is broadcast when a new hour begins.
type HourChangedEvent struct {
	Time time.Time
}

// DayChangedEvent is broadcast when a new day begins.
type DayChangedEvent struct {
	Time time.Time
}

// MonthChangedEvent is broadcast when a new month begins.
type MonthChangedEvent struct {
	Time time.Time
}

// YearChangedEvent is broadcast when a new year begins.
type YearChangedEvent struct {
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
