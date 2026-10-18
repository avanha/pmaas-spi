package bluetooth

import "testing"

// The System frame of a BlueCharm BC011 Pro, as a proxy captured it (the service data after the UUID).
var kkmSystemFrame = []byte{0x22, 0x0b, 0x67, 0xdd, 0x88, 0x00, 0x00, 0x3d, 0x36, 0x06, 0x3e}

func TestParseKKMSystem(t *testing.T) {
	got, ok := ParseKKMSystem(map[string][]byte{eddystoneUUID: kkmSystemFrame})
	if !ok {
		t.Fatal("not recognised")
	}

	// 0x0b is the model, 0x67 the battery (103%: over 100 on a good cell), then the address, and the firmware, 6.62.
	if got.Model != 11 || got.BatteryPercent != 103 || got.Address != "DD:88:00:00:3D:36" || got.FirmwareMajor != 6 || got.FirmwareMinor != 62 {
		t.Errorf("got %+v", got)
	}
}

func TestParseKKMSystem_RefusesWhatIsNotOne(t *testing.T) {
	for name, data := range map[string]map[string][]byte{
		"nothing":      nil,
		"another UUID": {"0000feab-0000-1000-8000-00805f9b34fb": kkmSystemFrame},
		"a TLM frame":  {eddystoneUUID: append([]byte{0x20}, kkmSystemFrame[1:]...)},
		"short":        {eddystoneUUID: kkmSystemFrame[:10]},
	} {
		if _, ok := ParseKKMSystem(data); ok {
			t.Errorf("%s was accepted", name)
		}
	}

	// And a TLM frame isn't mistaken for it the other way round.
	if _, ok := ParseEddystoneTLM(map[string][]byte{eddystoneUUID: kkmSystemFrame}); ok {
		t.Error("a System frame was read as TLM")
	}
}
