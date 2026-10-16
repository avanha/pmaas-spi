package automation

import (
	"strings"
	"testing"
	"time"
)

type fakeTrigger struct{}

func (fakeTrigger) Previous(time.Time) (time.Time, bool) { return time.Time{}, false }
func (fakeTrigger) Next(time.Time) (time.Time, bool)     { return time.Time{}, false }
func (fakeTrigger) Description() string                  { return "never" }

type fake struct {
	id, name string
	trigger  Trigger
	catchUp  time.Duration
}

func (f fake) ID() string             { return f.id }
func (f fake) Name() string           { return f.name }
func (f fake) Description() string    { return "" }
func (f fake) Trigger() Trigger       { return f.trigger }
func (f fake) CatchUp() time.Duration { return f.catchUp }
func (f fake) Run() (string, error)   { return "", nil }

func TestValidate(t *testing.T) {
	good := fake{id: "lights-on", name: "Turn on the lights", trigger: fakeTrigger{}, catchUp: time.Hour}

	if err := Validate(good); err != nil {
		t.Errorf("a good automation: %v", err)
	}

	zeroCatchUp := good
	zeroCatchUp.catchUp = 0

	if err := Validate(zeroCatchUp); err != nil {
		t.Errorf("a catch-up of zero is allowed: %v", err)
	}

	for name, c := range map[string]struct {
		automation Automation
		wants      []string
	}{
		"nil":        {nil, []string{"nil"}},
		"no id":      {fake{name: "n", trigger: fakeTrigger{}}, []string{"no ID"}},
		"no name":    {fake{id: "i", trigger: fakeTrigger{}}, []string{"no name"}},
		"no trigger": {fake{id: "i", name: "n"}, []string{"no trigger"}},
		"negative":   {fake{id: "i", name: "n", trigger: fakeTrigger{}, catchUp: -time.Minute}, []string{"negative"}},
		"several":    {fake{}, []string{"no ID", "no name", "no trigger"}},
	} {
		err := Validate(c.automation)
		if err == nil {
			t.Errorf("%s: no error", name)
			continue
		}

		for _, want := range c.wants {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: %q doesn't mention %q", name, err, want)
			}
		}
	}
}
