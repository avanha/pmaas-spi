// Package format says how long ago, or how long until, a time is, for pages: the plugins' templates all want it, and said it in
// several ways of their own. Every function takes the time from the clock of the machine it is on.
package format

import (
	"fmt"
	"time"
)

// Never is what a zero time reads as: something that has not happened.
const Never = "never"

// Ago says how long ago a time was, in its two most significant units: "5m ago", "2h 5m ago", "3d 4h ago". A time less than a minute
// ago, or still to come (the clocks of a device and this server needn't agree), reads as "just now", and a zero time as "never".
func Ago(t time.Time) string {
	if t.IsZero() {
		return Never
	}

	return span(time.Since(t), "just now") + suffixUnlessZero(time.Since(t), " ago")
}

// AgoSeconds is Ago down to the second for the first minute ("12s ago"), for what is a matter of seconds, such as whether data is
// flowing. A time still to come reads as "just now".
func AgoSeconds(t time.Time) string {
	if t.IsZero() {
		return Never
	}

	elapsed := time.Since(t).Truncate(time.Second)

	switch {
	case elapsed < 0:
		return "just now"
	case elapsed < time.Minute:
		return fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
	default:
		return Ago(t)
	}
}

// Until says how long until a time, in the same way: "in 5m", "in 2h 5m", "in 3d 4h". A time less than a minute away, or past, reads
// as "now", and a zero time as "never".
func Until(t time.Time) string {
	if t.IsZero() {
		return Never
	}

	remaining := time.Until(t)
	if text := span(remaining, ""); text != "" {
		return "in " + text
	}

	return "now"
}

// Remaining says how far a time is from now, either way: "in 6d 23h" for one still to come, "2h 5m ago" for one that has passed,
// with "< 1m" for less than a minute.
func Remaining(t time.Time) string {
	remaining := time.Until(t)

	if remaining >= 0 {
		return "in " + Span(remaining)
	}

	return Span(-remaining) + " ago"
}

// Span says how long a duration is in its two most significant units: "5m", "2h 5m", "3d 4h", or "< 1m".
func Span(d time.Duration) string {
	if text := span(d, ""); text != "" {
		return text
	}

	return "< 1m"
}

// Terse is Ago in the fewest characters, to the nearest unit and for a quick look at how stale a reading is: "< 30s", "< 1m", "5m",
// "3h" (to two days), "3d". A time still to come reads as "< 30s", and a zero time as "never".
func Terse(t time.Time) string {
	if t.IsZero() {
		return Never
	}

	elapsed := time.Since(t).Truncate(time.Second)

	switch {
	case elapsed < 30*time.Second:
		return "< 30s"
	case elapsed < time.Minute:
		return "< 1m"
	}

	elapsed = elapsed.Truncate(time.Minute)

	if elapsed < time.Hour {
		return fmt.Sprintf("%dm", int(elapsed.Minutes()))
	}

	elapsed = elapsed.Truncate(time.Hour)

	if elapsed < 48*time.Hour {
		return fmt.Sprintf("%dh", int(elapsed.Hours()))
	}

	return fmt.Sprintf("%dd", int(elapsed.Hours())/24)
}

// span is the duration in its two most significant units, or zero text if it is less than a minute (or negative).
func span(d time.Duration, zero string) string {
	d = d.Truncate(time.Minute)

	switch {
	case d <= 0:
		return zero
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		return fmt.Sprintf("%dd %dh", int(d.Hours())/24, int(d.Hours())%24)
	}
}

// suffixUnlessZero is the suffix, unless the duration read as zero ("just now" needs none).
func suffixUnlessZero(d time.Duration, suffix string) string {
	if d.Truncate(time.Minute) <= 0 {
		return ""
	}

	return suffix
}
