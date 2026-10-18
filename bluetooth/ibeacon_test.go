package bluetooth

import (
	"testing"
)

// iBeacon frames from the BlueCharm BC011 Pro, captured from a proxy: the manufacturer data after the length bytes.
const (
	bluecharmSingle = "0201061aff4c000215426c7565436861726d426561636f6e770efe1355c9"
	kkmDouble       = "0201061aff4c0002157777772e6b6b6d636e2e636f6d00000600010001c9"
)

func ibeaconOf(t *testing.T, frame string) (IBeacon, bool) {
	t.Helper()

	fields, err := ParseAdvertisingData(unhex(t, frame))
	if err != nil {
		t.Fatal(err)
	}

	return ParseIBeacon(fields.ManufacturerData)
}

func TestParseIBeacon_TheCapturedFrames(t *testing.T) {
	single, ok := ibeaconOf(t, bluecharmSingle)
	if !ok || single.UUID != "426c7565-4368-6172-6d42-6561636f6e77" || single.Major != 3838 || single.Minor != 4949 || single.MeasuredPower != -55 {
		t.Errorf("%v %+v", ok, single)
	}

	double, ok := ibeaconOf(t, kkmDouble)
	if !ok || double.UUID != "7777772e-6b6b-6d63-6e2e-636f6d000006" || double.Major != 1 || double.Minor != 1 || double.MeasuredPower != -55 {
		t.Errorf("%v %+v", ok, double)
	}
}

func TestParseIBeacon_AFrameWithTheScanResponseStillParses(t *testing.T) {
	// What a proxy sends when it heard the scan response with the advertisement.
	frame := bluecharmSingle + "0916802067040efe13550d09424350726f5f323232363136"

	if got, ok := ibeaconOf(t, frame); !ok || got.Minor != 4949 {
		t.Errorf("%v %+v", ok, got)
	}
}

func TestParseIBeacon_ThingsThatAreNotIBeacons(t *testing.T) {
	for name, data := range map[string]map[uint16][]byte{
		"none":                     nil,
		"another company":          {0x0766: make([]byte, 23)},
		"apple, another kind":      {0x004c: append([]byte{0x10, 0x05}, make([]byte, 21)...)},
		"apple, right type, short": {0x004c: append([]byte{0x02, 0x15}, make([]byte, 10)...)},
		"apple, right type, long":  {0x004c: append([]byte{0x02, 0x15}, make([]byte, 22)...)},
		"apple, wrong length byte": {0x004c: append([]byte{0x02, 0x16}, make([]byte, 21)...)},
	} {
		if got, ok := ParseIBeacon(data); ok {
			t.Errorf("%s: %+v", name, got)
		}
	}
}

func TestParseIBeacon_TheMajorAndMinorAreBigEndianAndThePowerIsSigned(t *testing.T) {
	data := append([]byte{0x02, 0x15}, make([]byte, 16)...)
	data = append(data, 0x01, 0x02, 0x03, 0x04, 0xd0)

	got, ok := ParseIBeacon(map[uint16][]byte{0x004c: data})
	if !ok || got.Major != 0x0102 || got.Minor != 0x0304 || got.MeasuredPower != -48 {
		t.Errorf("%v %+v", ok, got)
	}
}

func TestIBeaconMatch(t *testing.T) {
	frame := IBeacon{UUID: "426c7565-4368-6172-6d42-6561636f6e77", Major: 3838, Minor: 4949}

	for name, c := range map[string]struct {
		match IBeaconMatch
		want  bool
	}{
		"the uuid":                   {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77"), true},
		"the uuid in upper case":     {MatchIBeacon("426C7565-4368-6172-6D42-6561636F6E77"), true},
		"another uuid":               {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e73"), false},
		"the uuid and major":         {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77").WithMajor(3838), true},
		"the uuid and another major": {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77").WithMajor(1), false},
		"the uuid and minor":         {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77").WithMinor(4949), true},
		"the uuid and another minor": {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77").WithMinor(1), false},
		"all three":                  {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77").WithMajor(3838).WithMinor(4949), true},
		"a major of zero is a major": {MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77").WithMajor(0), false},
	} {
		if got := c.match.Matches(frame); got != c.want {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}

	if got := MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77").WithMajor(1).WithMinor(2).String(); got != "426c7565-4368-6172-6d42-6561636f6e77 major 1 minor 2" {
		t.Errorf("string %q", got)
	}
}

func TestMatchIBeacon_AnythingThatIsNotAUUIDPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic")
		}
	}()

	MatchIBeacon("not a uuid")
}

func TestIBeaconMatch_WithMajorDoesNotChangeTheMatchItWasMadeFrom(t *testing.T) {
	base := MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77")
	_ = base.WithMajor(1)

	if base.Major != nil {
		t.Error("the original was changed")
	}
}
