package automation

import (
	"strings"
	"testing"

	"github.com/avanha/pmaas-spi/events"
)

type fakeEventTrigger struct{}

func (fakeEventTrigger) Matches(*events.EventInfo) bool { return true }
func (fakeEventTrigger) Description() string            { return "always" }

type fakeEventAutomation struct {
	id, name string
	trigger  EventTrigger
}

func (f fakeEventAutomation) ID() string                 { return f.id }
func (f fakeEventAutomation) Name() string               { return f.name }
func (f fakeEventAutomation) Description() string        { return "" }
func (f fakeEventAutomation) EventTrigger() EventTrigger { return f.trigger }
func (f fakeEventAutomation) Run() (string, error)       { return "", nil }

func TestValidateEvent(t *testing.T) {
	if err := ValidateEvent(fakeEventAutomation{id: "lights", name: "Lights", trigger: fakeEventTrigger{}}); err != nil {
		t.Errorf("a good automation: %v", err)
	}

	for name, c := range map[string]struct {
		automation EventAutomation
		wants      []string
	}{
		"nil":        {nil, []string{"nil"}},
		"no id":      {fakeEventAutomation{name: "n", trigger: fakeEventTrigger{}}, []string{"no ID"}},
		"no name":    {fakeEventAutomation{id: "i", trigger: fakeEventTrigger{}}, []string{"no name"}},
		"no trigger": {fakeEventAutomation{id: "i", name: "n"}, []string{"no trigger"}},
		"all of it":  {fakeEventAutomation{}, []string{"no ID", "no name", "no trigger"}},
	} {
		err := ValidateEvent(c.automation)
		if err == nil {
			t.Errorf("%s: no error", name)

			continue
		}

		for _, want := range c.wants {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: %q does not say %q", name, err, want)
			}
		}
	}
}
