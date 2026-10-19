package format

import (
	"testing"
	"time"
)

func ago(d time.Duration) time.Time { return time.Now().Add(-d) }

func TestAgo(t *testing.T) {
	for name, c := range map[string]struct {
		t    time.Time
		want string
	}{
		"zero":                   {time.Time{}, "never"},
		"now":                    {time.Now(), "just now"},
		"a little in the future": {time.Now().Add(30 * time.Second), "just now"},
		"a long way ahead":       {time.Now().Add(3 * time.Hour), "just now"},
		"seconds":                {ago(40 * time.Second), "just now"},
		"minutes":                {ago(5*time.Minute + 10*time.Second), "5m ago"},
		"an hour":                {ago(time.Hour + time.Second), "1h 0m ago"},
		"hours":                  {ago(3*time.Hour + 20*time.Minute + 10*time.Second), "3h 20m ago"},
		"days":                   {ago(2*24*time.Hour + 5*time.Hour + time.Minute), "2d 5h ago"},
	} {
		if got := Ago(c.t); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestAgoSeconds(t *testing.T) {
	for name, c := range map[string]struct {
		t    time.Time
		want string
	}{
		"zero":     {time.Time{}, "never"},
		"future":   {time.Now().Add(time.Minute), "just now"},
		"seconds":  {ago(12*time.Second + 300*time.Millisecond), "12s ago"},
		"a minute": {ago(65 * time.Second), "1m ago"},
		"hours":    {ago(2*time.Hour + 5*time.Minute + time.Second), "2h 5m ago"},
		"days":     {ago(3*24*time.Hour + 4*time.Hour + time.Minute), "3d 4h ago"},
	} {
		if got := AgoSeconds(c.t); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestUntil(t *testing.T) {
	for name, c := range map[string]struct {
		t    time.Time
		want string
	}{
		"zero":    {time.Time{}, "never"},
		"past":    {time.Now().Add(-time.Hour), "now"},
		"seconds": {time.Now().Add(20 * time.Second), "now"},
		"minutes": {time.Now().Add(5*time.Minute + 20*time.Second), "in 5m"},
		"a day":   {time.Now().Add(25*time.Hour + 30*time.Second), "in 1d 1h"},
	} {
		if got := Until(c.t); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestSpanAndRemaining(t *testing.T) {
	for _, c := range []struct {
		d    time.Duration
		want string
	}{
		{0, "< 1m"},
		{59 * time.Second, "< 1m"},
		{5 * time.Minute, "5m"},
		{59*time.Minute + 59*time.Second, "59m"},
		{time.Hour, "1h 0m"},
		{5*time.Hour + 12*time.Minute, "5h 12m"},
		{23*time.Hour + 59*time.Minute, "23h 59m"},
		{24 * time.Hour, "1d 0h"},
		{6*24*time.Hour + 23*time.Hour + 30*time.Minute, "6d 23h"},
	} {
		if got := Span(c.d); got != c.want {
			t.Errorf("Span(%v) = %q, want %q", c.d, got, c.want)
		}
	}

	// The margins keep these from being flaky: time passes between building the value and formatting it.
	if got := Remaining(time.Now().Add(3*time.Hour + 30*time.Second)); got != "in 3h 0m" {
		t.Errorf("future: got %q", got)
	}

	if got := Remaining(time.Now().Add(-(2*time.Hour + 5*time.Minute + 30*time.Second))); got != "2h 5m ago" {
		t.Errorf("past: got %q", got)
	}

	if got := Remaining(time.Now().Add(20 * time.Second)); got != "in < 1m" {
		t.Errorf("soon: got %q", got)
	}
}

func TestTerse(t *testing.T) {
	for name, c := range map[string]struct {
		t    time.Time
		want string
	}{
		"zero":     {time.Time{}, "never"},
		"recent":   {ago(5 * time.Second), "< 30s"},
		"future":   {time.Now().Add(time.Minute), "< 30s"},
		"a bit":    {ago(45 * time.Second), "< 1m"},
		"minutes":  {ago(5*time.Minute + 10*time.Second), "5m"},
		"hours":    {ago(3*time.Hour + time.Minute), "3h"},
		"47 hours": {ago(47*time.Hour + 30*time.Minute), "47h"},
		"days":     {ago(72*time.Hour + time.Minute), "3d"},
	} {
		if got := Terse(c.t); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}
