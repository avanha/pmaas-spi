package clock

import (
	"testing"

	"github.com/avanha/pmaas-spi/events"
)

func TestIsClockEvent(t *testing.T) {
	for name, event := range map[string]any{
		"minute": MinuteChangedEvent{}, "hour": HourChangedEvent{}, "day": DayChangedEvent{},
		"month": MonthChangedEvent{}, "year": YearChangedEvent{},
	} {
		if !IsClockEvent(&events.EventInfo{Event: event}) {
			t.Errorf("%s change isn't a clock event", name)
		}
	}

	for name, event := range map[string]any{
		"another entity's": events.EntityStateChangedEvent{}, "a string": "minute", "nil": nil,
	} {
		if IsClockEvent(&events.EventInfo{Event: event}) {
			t.Errorf("%s event is a clock event", name)
		}
	}
}
