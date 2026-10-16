// Package alert is how plugins and the server tell the user about something that needs their attention.
//
// A producer raises an alert by broadcasting a RaisedEvent, and clears it by broadcasting a ClearedEvent,
// using the Raise and Clear helpers. It never needs to know who is listening, or whether anyone is: the
// alert console plugin collects them for display, and another plugin could subscribe to the same events
// to send an email or a push notification. A producer works the same whether or not any of them is
// installed.
//
// Alerts are state, not notifications. An alert is identified by its Source and Key, and stays active
// until its producer clears it: raising the same Source and Key again updates the existing alert rather
// than adding another, so a producer that re-evaluates its condition regularly can simply raise it every
// time, and an alert that escalates (warning, then critical) is one alert whose severity changed.
package alert

import (
	"fmt"
	"time"

	"github.com/avanha/pmaas-spi/events"
)

// Severity is how urgent an alert is. The zero value means the producer didn't say, and is treated as
// SeverityInfo.
type Severity int

const (
	// SeverityInfo is something worth knowing, which needs nothing doing about it.
	SeverityInfo Severity = iota + 1
	// SeverityWarning is something that will need attention, but not yet.
	SeverityWarning
	// SeverityCritical is something that needs attention now.
	SeverityCritical
)

// Normalize returns s, or SeverityInfo if s isn't one of the defined severities.
func (s Severity) Normalize() Severity {
	switch s {
	case SeverityInfo, SeverityWarning, SeverityCritical:
		return s
	default:
		return SeverityInfo
	}
}

func (s Severity) String() string {
	switch s.Normalize() {
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	default:
		return "info"
	}
}

// RaisedEvent is broadcast to raise an alert, or to update one that's already active.
type RaisedEvent struct {
	// Source is who raised the alert: a plugin's ShortName, or "core" for the server itself.
	Source string

	// Key identifies the alert within Source, and must stay the same for as long as the alert is about
	// the same thing, e.g. "refresh-token". Raising an alert whose Source and Key are already active
	// updates it.
	Key string

	Severity Severity

	// Title is a short headline, and Message the detail: what's wrong and what to do about it.
	Title   string
	Message string
}

// ClearedEvent is broadcast when whatever an alert was about is resolved. Clearing an alert that isn't
// active does nothing.
type ClearedEvent struct {
	Source string
	Key    string
}

// SummaryEvent is broadcast by whatever keeps track of which alerts the user has dealt with (the alert console),
// each time that changes, so that something else can show the state at a glance without keeping the alerts
// itself, such as an indicator in a page header. It's about what still needs the user's attention: an alert
// the user has acknowledged isn't counted, until it escalates.
//
// It isn't a RaisedEvent or ClearedEvent, and EventPredicate doesn't match it, so a plugin that listens for
// alerts is never handed one.
type SummaryEvent struct {
	// Unacknowledged is how many active alerts the user hasn't acknowledged.
	Unacknowledged int

	// MostSevere is the highest severity among those. It's zero when Unacknowledged is zero.
	MostSevere Severity
}

// Broadcaster is what's needed to raise and clear alerts, which spi.IPMAASContainer provides.
type Broadcaster interface {
	BroadcastEvent(entityEventId string, event any) error
}

// Raise raises, or updates, the alert identified by source and key.
//
// Events are broadcast from an entity, and an alert doesn't come from one, so source stands in for the
// entity id.
func Raise(broadcaster Broadcaster, source string, key string, severity Severity, title string, message string) error {
	return broadcaster.BroadcastEvent(source, RaisedEvent{
		Source:   source,
		Key:      key,
		Severity: severity,
		Title:    title,
		Message:  message,
	})
}

// Clear clears the alert identified by source and key.
func Clear(broadcaster Broadcaster, source string, key string) error {
	return broadcaster.BroadcastEvent(source, ClearedEvent{Source: source, Key: key})
}

// EventPredicate is an events.EventPredicate matching the events in this package, for a plugin
// registering to receive alerts.
func EventPredicate(eventInfo *events.EventInfo) bool {
	switch eventInfo.Event.(type) {
	case RaisedEvent, ClearedEvent:
		return true
	default:
		return false
	}
}

// PublishSummary broadcasts a SummaryEvent. source stands in for the entity id, as it does for Raise.
func PublishSummary(broadcaster Broadcaster, source string, unacknowledged int, mostSevere Severity) error {
	if unacknowledged <= 0 {
		unacknowledged, mostSevere = 0, 0
	}

	return broadcaster.BroadcastEvent(source, SummaryEvent{Unacknowledged: unacknowledged, MostSevere: mostSevere})
}

// SummaryPredicate is an events.EventPredicate matching SummaryEvents, for a plugin that shows them.
func SummaryPredicate(eventInfo *events.EventInfo) bool {
	_, ok := eventInfo.Event.(SummaryEvent)

	return ok
}

// SeverityFunc says how severe an alert is at a time. A producer that holds an alert for as long as a condition holds can
// ask it every time it looks, and raise the alert again, with the new severity, when the answer changes: the alert
// escalates, or eases, as the day goes on. Always, ByTimeOfDay and the builders that go with the automations library
// make them.
type SeverityFunc func(now time.Time) Severity

// Always is a severity that doesn't depend on the time.
func Always(severity Severity) SeverityFunc {
	return func(time.Time) Severity { return severity }
}

// ByTimeOfDay is inside from the first moment of the daily window that begins at fromHour:fromMinute until the moment
// it ends at toHour:toMinute, which is not in it, and outside the rest of the day. The window can cross midnight:
// ByTimeOfDay(SeverityWarning, 7, 0, 23, 0, SeverityCritical) is a warning from 7am until 11pm and critical overnight, and
// ByTimeOfDay(SeverityCritical, 23, 0, 7, 0, SeverityWarning) is the same the other way round. It goes by the time on
// the clock in the zone of the time it is asked about, so on a day when the clocks change the window is the same hours on
// the wall.
//
// It panics if the window isn't one, because the hours are not of a day or because it begins when it ends, since that is
// either no time or all of it and no one means either, like regexp.MustCompile does for a pattern that can't be.
func ByTimeOfDay(inside Severity, fromHour int, fromMinute int, toHour int, toMinute int, outside Severity) SeverityFunc {
	if !validTimeOfDay(fromHour, fromMinute) || !validTimeOfDay(toHour, toMinute) {
		panic(fmt.Sprintf("alert.ByTimeOfDay: %d:%02d to %d:%02d is not a window of the day", fromHour, fromMinute, toHour, toMinute))
	}

	from, to := fromHour*60+fromMinute, toHour*60+toMinute
	if from == to {
		panic(fmt.Sprintf("alert.ByTimeOfDay: the window from %d:%02d to %d:%02d is empty or all day", fromHour, fromMinute, toHour, toMinute))
	}

	return func(now time.Time) Severity {
		minute := now.Hour()*60 + now.Minute()

		var in bool
		if from < to {
			in = minute >= from && minute < to
		} else {
			in = minute >= from || minute < to
		}

		if in {
			return inside
		}

		return outside
	}
}

func validTimeOfDay(hour int, minute int) bool {
	return hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}
