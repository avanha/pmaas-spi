package alert

import (
	"fmt"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

// OfflineAlert is the setting a plugin offers for a device it looks after: whether to raise an alert when the device
// goes offline, how severe, and how long it has to have been offline first. The zero value is no alert, so a device
// that is not configured for one never has one. Make one with WhenOffline.
type OfflineAlert struct {
	enabled  bool
	after    time.Duration
	severity SeverityFunc
}

// WhenOffline raises an alert once the device has been offline for at least after, and clears it when the device is back.
// A time of zero is as soon as it is noticed. severity says how severe the alert is, and is asked every time the device
// is looked at, so it can escalate or ease through the day (see Always and ByTimeOfDay); with nil it is a warning.
//
// It panics for a negative time, like regexp.MustCompile does for a pattern that can't be: it is configuration written in
// the program.
func WhenOffline(after time.Duration, severity SeverityFunc) OfflineAlert {
	if after < 0 {
		panic(fmt.Sprintf("alert.WhenOffline: the time to be offline before alerting can't be negative, not %v", after))
	}

	if severity == nil {
		severity = Always(SeverityWarning)
	}

	return OfflineAlert{enabled: true, after: after, severity: severity}
}

// Enabled is whether an alert was asked for.
func (o OfflineAlert) Enabled() bool {
	return o.enabled
}

// After is how long the device has to be offline before the alert is raised.
func (o OfflineAlert) After() time.Duration {
	return o.after
}

// Watch makes what a plugin keeps for one device to carry the setting out: source is the plugin, key says which alert it is
// among the plugin's, and name is the device as people know it. It returns nil if no alert was asked for, and a nil watch
// does nothing when checked, so a plugin can check every device the same.
func (o OfflineAlert) Watch(source string, key string, name string) *OfflineWatch {
	if !o.enabled {
		return nil
	}

	return &OfflineWatch{policy: o, source: source, key: key, name: name}
}

// OfflineText words an offline alert: its title and its message, given when the device went offline and how long ago that is.
type OfflineText func(since time.Time, offline time.Duration) (title string, message string)

// OfflineWatch holds one device's offline alert. It remembers what it has raised, so that only a change is sent and it
// can be checked as often as is wanted. Not safe to use from more than one goroutine.
type OfflineWatch struct {
	policy   OfflineAlert
	source   string
	key      string
	name     string
	text     OfflineText
	raised   bool
	severity Severity
}

// WithText has the alert worded by text instead of in the usual way ("Offline: name", and how long it has been), for
// something that is not quite a device going offline, say a host that is unreachable. It returns the watch, and does
// nothing to a nil one.
func (w *OfflineWatch) WithText(text OfflineText) *OfflineWatch {
	if w != nil {
		w.text = text
	}

	return w
}

// Check looks at the device at a time now, given whether it is connected and, if not, since when.
//
// Once it has been offline for the setting's time the alert is raised at the severity for the time, and raised again if that
// changes. When it is online again the alert is cleared. When it is not known whether it is online, or it has not been
// offline long enough, nothing is done: an alert once raised isn't cleared by a look that tells nothing.
func (w *OfflineWatch) Check(broadcaster Broadcaster, connectivity environment.Connectivity, offlineSince time.Time, now time.Time) error {
	if w == nil {
		return nil
	}

	switch connectivity {
	case environment.ConnectivityOnline:
		if !w.raised {
			return nil
		}

		if err := Clear(broadcaster, w.source, w.key); err != nil {
			return err
		}

		w.raised = false

		return nil
	case environment.ConnectivityOffline:
		offline := now.Sub(offlineSince)
		if offline < w.policy.after {
			return nil
		}

		severity := w.policy.severity(now).Normalize()
		if w.raised && w.severity == severity {
			return nil
		}

		title, message := "Offline: "+w.name, fmt.Sprintf("%s has been offline since %s, %s ago.",
			w.name, offlineSince.Format("2006-01-02 3:04 PM"), offline.Round(time.Second))

		if w.text != nil {
			title, message = w.text(offlineSince, offline)
		}

		if err := Raise(broadcaster, w.source, w.key, severity, title, message); err != nil {
			return err
		}

		w.raised, w.severity = true, severity

		return nil
	default:
		return nil
	}
}
