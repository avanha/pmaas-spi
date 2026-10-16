package alert

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

var epoch = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func TestOfflineAlert_IsOffUntilAskedFor(t *testing.T) {
	var zero OfflineAlert

	if zero.Enabled() || zero.Watch("s", "k", "n") != nil {
		t.Fatal("the zero value should be no alert")
	}

	// A nil watch can be checked.
	if err := zero.Watch("s", "k", "n").Check(&recordingBroadcaster{}, environment.ConnectivityOffline, epoch, epoch.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
}

func TestWhenOffline_RejectsANegativeTime(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic")
		}
	}()

	WhenOffline(-time.Second, nil)
}

func TestOfflineWatch_WaitsForTheTimeThenRaisesOnceAndClearsWhenBack(t *testing.T) {
	b := &recordingBroadcaster{}
	w := WhenOffline(5*time.Minute, nil).Watch("fakes", "offline:door", "Garage Door")

	check := func(c environment.Connectivity, at time.Duration) {
		t.Helper()

		if err := w.Check(b, c, epoch, epoch.Add(at)); err != nil {
			t.Fatal(err)
		}
	}

	check(environment.ConnectivityOffline, 4*time.Minute+59*time.Second)

	if len(b.events) != 0 {
		t.Fatalf("raised too early: %+v", b.events)
	}

	check(environment.ConnectivityOffline, 5*time.Minute)
	check(environment.ConnectivityOffline, 6*time.Minute)

	want := RaisedEvent{
		Source: "fakes", Key: "offline:door", Severity: SeverityWarning, Title: "Offline: Garage Door",
		Message: "Garage Door has been offline since 2026-03-10 12:00 PM, 5m0s ago.",
	}
	if len(b.events) != 1 || !reflect.DeepEqual(b.events[0], want) {
		t.Fatalf("got %+v", b.events)
	}

	check(environment.ConnectivityUnknown, 7*time.Minute)

	if len(b.events) != 1 {
		t.Fatalf("an unknown state changed the alert: %+v", b.events)
	}

	check(environment.ConnectivityOnline, 8*time.Minute)
	check(environment.ConnectivityOnline, 9*time.Minute)

	if len(b.events) != 2 || !reflect.DeepEqual(b.events[1], ClearedEvent{Source: "fakes", Key: "offline:door"}) {
		t.Fatalf("got %+v", b.events)
	}
}

func TestOfflineWatch_NoTimeMeansAtOnce(t *testing.T) {
	b := &recordingBroadcaster{}
	w := WhenOffline(0, Always(SeverityCritical)).Watch("s", "k", "n")

	if err := w.Check(b, environment.ConnectivityOffline, epoch, epoch); err != nil {
		t.Fatal(err)
	}

	if len(b.events) != 1 || b.events[0].(RaisedEvent).Severity != SeverityCritical {
		t.Fatalf("got %+v", b.events)
	}
}

func TestOfflineWatch_EscalatesWhenTheSeverityChanges(t *testing.T) {
	b := &recordingBroadcaster{}
	w := WhenOffline(0, ByTimeOfDay(SeverityWarning, 7, 0, 23, 0, SeverityCritical)).Watch("s", "k", "n")
	since := time.Date(2026, 3, 10, 22, 0, 0, 0, time.UTC)

	for _, at := range []time.Time{since.Add(30 * time.Minute), since.Add(45 * time.Minute), since.Add(70 * time.Minute)} {
		if err := w.Check(b, environment.ConnectivityOffline, since, at); err != nil {
			t.Fatal(err)
		}
	}

	var severities []Severity
	for _, e := range b.events {
		severities = append(severities, e.(RaisedEvent).Severity)
	}

	if !reflect.DeepEqual(severities, []Severity{SeverityWarning, SeverityCritical}) {
		t.Fatalf("got %v", severities)
	}
}

func TestOfflineWatch_ARaiseThatFailsIsTriedAgain(t *testing.T) {
	b := &recordingBroadcaster{err: errors.New("no")}
	w := WhenOffline(0, nil).Watch("s", "k", "n")

	if err := w.Check(b, environment.ConnectivityOffline, epoch, epoch); err == nil {
		t.Fatal("no error")
	}

	b.err = nil

	if err := w.Check(b, environment.ConnectivityOffline, epoch, epoch); err != nil {
		t.Fatal(err)
	}

	if len(b.events) != 2 {
		t.Fatalf("expected the retry to be sent: %+v", b.events)
	}
}

func TestOfflineWatch_WithTextWordsTheAlert_AndANilWatchIgnoresIt(t *testing.T) {
	b := &recordingBroadcaster{}
	w := WhenOffline(0, nil).Watch("s", "k", "n").WithText(func(since time.Time, offline time.Duration) (string, string) {
		return "Host down", "down since " + since.Format("15:04") + " for " + offline.String()
	})

	if err := w.Check(b, environment.ConnectivityOffline, epoch, epoch.Add(3*time.Minute)); err != nil {
		t.Fatal(err)
	}

	got := b.events[0].(RaisedEvent)
	if got.Title != "Host down" || got.Message != "down since 12:00 for 3m0s" {
		t.Errorf("got %+v", got)
	}

	var none *OfflineWatch
	if none.WithText(nil) != nil {
		t.Error("a nil watch became something")
	}
}
