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
