package logging

import (
	"bytes"
	"log/slog"
	"testing"
)

type provider struct{ names []string }

func (p *provider) Logger(names ...string) *slog.Logger {
	p.names = names
	return slog.New(slog.DiscardHandler)
}
func (p *provider) LogController() Controller { return nil }

func TestFor_AsksAContainerThatHandsOutLoggers(t *testing.T) {
	p := &provider{}

	if For(p, "broker") == nil || len(p.names) != 1 || p.names[0] != "broker" {
		t.Errorf("asked for %v", p.names)
	}
}

func TestFor_FallsBackToTheDefaultLoggerForAnyOtherContainer(t *testing.T) {
	var out bytes.Buffer

	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&out, nil)))

	defer slog.SetDefault(old)

	For(struct{}{}).Warn("hello")

	if !bytes.Contains(out.Bytes(), []byte("hello")) {
		t.Errorf("logged %q", out.String())
	}

	if ControllerOf(struct{}{}) != nil {
		t.Error("a controller from nothing")
	}
}

func TestValid(t *testing.T) {
	for name, want := range map[string]bool{
		"plugin": true, "plugin.esphome": true, "": false, "a..b": false, ".a": false, "a.": false, "a b": false, "a=b": false,
	} {
		if Valid(name) != want {
			t.Errorf("Valid(%q) = %v", name, !want)
		}
	}
}

func TestDepthAndOrder(t *testing.T) {
	if Depth("") != 0 || Depth("a") != 1 || Depth("a.b") != 2 {
		t.Error("depths")
	}

	c := []Component{{Name: "b"}, {Name: "a.b"}, {Name: ""}, {Name: "a"}}
	SortByName(c)

	for i, want := range []string{"", "a", "a.b", "b"} {
		if c[i].Name != want {
			t.Fatalf("%d: %q", i, c[i].Name)
		}
	}
}

func TestValidLevel(t *testing.T) {
	if !ValidLevel(slog.LevelWarn) || ValidLevel(slog.LevelWarn+1) || ValidLevel(slog.Level(-8)) {
		t.Error("levels")
	}
}
