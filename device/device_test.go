package device

import "testing"

func TestSwitchState_ZeroValueIsUnknownNotOff(t *testing.T) {
	var zero SwitchState

	if zero != SwitchStateUnknown || zero.String() != "Unknown" {
		t.Fatalf("zero value is %v", zero)
	}

	if SwitchStateOn.String() != "On" || SwitchStateOff.String() != "Off" {
		t.Fatal("unexpected names")
	}
}

func TestDoorPosition_ZeroValueIsUnknownNotClosed(t *testing.T) {
	var zero DoorPosition

	if zero != DoorPositionUnknown || zero.String() != "Unknown" {
		t.Fatalf("zero value is %v", zero)
	}
}

func TestEntityTypes(t *testing.T) {
	for _, c := range []struct {
		name string
		got  string
	}{
		{"ISwitch", ISwitchType.String()},
		{"IPowerMonitor", IPowerMonitorType.String()},
		{"IDoorSensor", IDoorSensorType.String()},
	} {
		if c.got != "device."+c.name {
			t.Errorf("%s registered as %s", c.name, c.got)
		}
	}
}

func TestSwitchState_OnlyOnAndOffCanBeAskedFor(t *testing.T) {
	for state, want := range map[SwitchState]bool{
		SwitchStateOn: true, SwitchStateOff: true, SwitchStateUnknown: false, SwitchState(99): false, SwitchState(-1): false,
	} {
		if got := state.IsSettable(); got != want {
			t.Errorf("%d.IsSettable() = %v, want %v", state, got, want)
		}
	}
}

func TestButton(t *testing.T) {
	for press, want := range map[ButtonPress]string{
		PressUnknown: "unknown", PressSingle: "single", PressDouble: "double", PressTriple: "triple", PressLong: "long", ButtonPress(99): "unknown",
	} {
		if press.String() != want {
			t.Errorf("%d: %q, want %q", press, press.String(), want)
		}
	}

	if IButtonType.String() != "device.IButton" {
		t.Errorf("type %s", IButtonType)
	}

	var zero Button
	if zero.Presses != 0 || zero.LastPress != PressUnknown || !zero.LastPressTime.IsZero() {
		t.Error("a new button looks pressed")
	}
}
