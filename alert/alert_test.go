package alert

import (
	"errors"
	"reflect"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/avanha/pmaas-spi/events"
)

type recordingBroadcaster struct {
	ids    []string
	events []any
	err    error
}

func (r *recordingBroadcaster) BroadcastEvent(entityEventId string, event any) error {
	r.ids = append(r.ids, entityEventId)
	r.events = append(r.events, event)

	return r.err
}

func TestRaise_BroadcastsARaisedEventFromTheSource(t *testing.T) {
	b := &recordingBroadcaster{}

	if err := Raise(b, "nestthermostat", "refresh-token", SeverityWarning, "Title", "Message"); err != nil {
		t.Fatal(err)
	}

	want := RaisedEvent{
		Source: "nestthermostat", Key: "refresh-token", Severity: SeverityWarning, Title: "Title", Message: "Message",
	}

	if len(b.events) != 1 || !reflect.DeepEqual(b.events[0], want) || b.ids[0] != "nestthermostat" {
		t.Fatalf("unexpected broadcast: ids=%v events=%+v", b.ids, b.events)
	}
}

func TestClear_BroadcastsAClearedEventFromTheSource(t *testing.T) {
	b := &recordingBroadcaster{}

	if err := Clear(b, "nestthermostat", "refresh-token"); err != nil {
		t.Fatal(err)
	}

	want := ClearedEvent{Source: "nestthermostat", Key: "refresh-token"}

	if len(b.events) != 1 || !reflect.DeepEqual(b.events[0], want) || b.ids[0] != "nestthermostat" {
		t.Fatalf("unexpected broadcast: ids=%v events=%+v", b.ids, b.events)
	}
}

func TestRaiseAndClear_ReportBroadcastFailures(t *testing.T) {
	b := &recordingBroadcaster{err: errors.New("no longer accepting events")}

	if err := Raise(b, "s", "k", SeverityInfo, "t", "m"); err == nil {
		t.Error("Raise swallowed the error")
	}

	if err := Clear(b, "s", "k"); err == nil {
		t.Error("Clear swallowed the error")
	}
}

func TestEventPredicate(t *testing.T) {
	for name, c := range map[string]struct {
		event any
		want  bool
	}{
		"raised":         {RaisedEvent{}, true},
		"cleared":        {ClearedEvent{}, true},
		"pointer to it":  {&RaisedEvent{}, false},
		"another event":  {events.EntityRegisteredEvent{}, false},
		"not an event":   {"hello", false},
		"nothing at all": {nil, false},
	} {
		if got := EventPredicate(&events.EventInfo{Event: c.event}); got != c.want {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
}

func TestSeverity(t *testing.T) {
	for severity, want := range map[Severity]string{
		SeverityInfo: "info", SeverityWarning: "warning", SeverityCritical: "critical", 0: "info", 99: "info", -1: "info",
	} {
		if got := severity.String(); got != want {
			t.Errorf("%d: String() = %q, want %q", severity, got, want)
		}
	}

	if Severity(0).Normalize() != SeverityInfo || Severity(99).Normalize() != SeverityInfo ||
		SeverityCritical.Normalize() != SeverityCritical {
		t.Error("Normalize doesn't keep defined severities and default the rest to info")
	}

	if !(SeverityInfo < SeverityWarning && SeverityWarning < SeverityCritical) {
		t.Error("severities must order from least to most urgent, callers compare them")
	}
}

func TestPublishSummary_BroadcastsASummaryEventFromTheSource(t *testing.T) {
	b := &recordingBroadcaster{}

	if err := PublishSummary(b, "alertconsole", 2, SeverityCritical); err != nil {
		t.Fatal(err)
	}

	want := SummaryEvent{Unacknowledged: 2, MostSevere: SeverityCritical}

	if len(b.events) != 1 || !reflect.DeepEqual(b.events[0], want) || b.ids[0] != "alertconsole" {
		t.Fatalf("unexpected broadcast: ids=%v events=%+v", b.ids, b.events)
	}
}

func TestPublishSummary_NothingUnacknowledgedHasNoSeverity(t *testing.T) {
	b := &recordingBroadcaster{}

	if err := PublishSummary(b, "alertconsole", 0, SeverityWarning); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(b.events[0], SummaryEvent{}) {
		t.Fatalf("unexpected event %+v", b.events[0])
	}
}

func TestPublishSummary_ReturnsTheBroadcastError(t *testing.T) {
	want := errors.New("boom")

	if err := PublishSummary(&recordingBroadcaster{err: want}, "alertconsole", 1, SeverityInfo); !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestPredicates_SummaryIsNotAnAlertAndAnAlertIsNotASummary(t *testing.T) {
	summary := &events.EventInfo{Event: SummaryEvent{}}
	raised := &events.EventInfo{Event: RaisedEvent{}}

	if !SummaryPredicate(summary) || SummaryPredicate(raised) || SummaryPredicate(&events.EventInfo{Event: "x"}) {
		t.Error("SummaryPredicate matched the wrong events")
	}

	// A console that re-published its summary on receiving one would loop forever.
	if EventPredicate(summary) {
		t.Error("EventPredicate must not match a SummaryEvent")
	}
}

func TestAlways(t *testing.T) {
	f := Always(SeverityCritical)

	for _, at := range []time.Time{{}, time.Date(2026, 10, 16, 3, 0, 0, 0, time.UTC), time.Now()} {
		if got := f(at); got != SeverityCritical {
			t.Errorf("%v: %v", at, got)
		}
	}
}

func TestByTimeOfDay_ADayTimeWindow(t *testing.T) {
	f := ByTimeOfDay(SeverityWarning, 7, 0, 23, 0, SeverityCritical)

	at := func(hour, minute int) time.Time { return time.Date(2026, 10, 16, hour, minute, 30, 0, time.UTC) }

	for _, c := range []struct {
		hour, minute int
		want         Severity
	}{
		{0, 0, SeverityCritical}, {6, 59, SeverityCritical}, {7, 0, SeverityWarning}, {12, 0, SeverityWarning},
		{22, 59, SeverityWarning}, {23, 0, SeverityCritical}, {23, 59, SeverityCritical},
	} {
		if got := f(at(c.hour, c.minute)); got != c.want {
			t.Errorf("%02d:%02d: %v, want %v", c.hour, c.minute, got, c.want)
		}
	}
}

func TestByTimeOfDay_BoundariesWithMinutes(t *testing.T) {
	f := ByTimeOfDay(SeverityWarning, 7, 30, 22, 45, SeverityCritical)

	at := func(hour, minute int) time.Time { return time.Date(2026, 10, 16, hour, minute, 0, 0, time.UTC) }

	for _, c := range []struct {
		hour, minute int
		want         Severity
	}{
		{7, 29, SeverityCritical}, {7, 30, SeverityWarning}, {22, 44, SeverityWarning}, {22, 45, SeverityCritical},
	} {
		if got := f(at(c.hour, c.minute)); got != c.want {
			t.Errorf("%02d:%02d: %v, want %v", c.hour, c.minute, got, c.want)
		}
	}
}

func TestByTimeOfDay_AWindowThatCrossesMidnight(t *testing.T) {
	f := ByTimeOfDay(SeverityCritical, 23, 0, 7, 0, SeverityWarning)

	at := func(hour, minute int) time.Time { return time.Date(2026, 10, 16, hour, minute, 0, 0, time.UTC) }

	for _, c := range []struct {
		hour, minute int
		want         Severity
	}{
		{22, 59, SeverityWarning}, {23, 0, SeverityCritical}, {23, 59, SeverityCritical}, {0, 0, SeverityCritical},
		{6, 59, SeverityCritical}, {7, 0, SeverityWarning}, {12, 0, SeverityWarning},
	} {
		if got := f(at(c.hour, c.minute)); got != c.want {
			t.Errorf("%02d:%02d: %v, want %v", c.hour, c.minute, got, c.want)
		}
	}
}

func TestByTimeOfDay_GoesByTheClockOnTheWallInTheZoneItIsGiven(t *testing.T) {
	denver, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Skip("no time zone data")
	}

	f := ByTimeOfDay(SeverityWarning, 7, 0, 23, 0, SeverityCritical)

	// 11:30pm Mountain time is 5:30am UTC the next day.
	late := time.Date(2026, 10, 16, 23, 30, 0, 0, denver)

	if got := f(late); got != SeverityCritical {
		t.Errorf("11:30pm in Denver: %v", got)
	}

	if got := f(late.UTC()); got != SeverityCritical { // 05:30 UTC: also before 7
		t.Errorf("the same moment in UTC: %v", got)
	}

	noon := time.Date(2026, 10, 16, 12, 0, 0, 0, denver)

	if got := f(noon); got != SeverityWarning {
		t.Errorf("noon in Denver: %v", got)
	}
}

func TestByTimeOfDay_PanicsForAWindowThatIsNotOne(t *testing.T) {
	for name, c := range map[string][4]int{
		"hour too big":   {24, 0, 7, 0},
		"minute too big": {7, 60, 9, 0},
		"negative":       {-1, 0, 7, 0},
		"empty":          {7, 0, 7, 0},
		"to is invalid":  {7, 0, 8, 61},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: didn't panic", name)
				}
			}()

			ByTimeOfDay(SeverityWarning, c[0], c[1], c[2], c[3], SeverityCritical)
		}()
	}
}
