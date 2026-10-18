package bluetooth

import (
	"testing"
	"time"
)

func tlmFrame(mutate func([]byte)) map[string][]byte {
	data := []byte{0x20, 0x00, 0x0b, 0xb8, 0x18, 0x80, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0xe8}
	if mutate != nil {
		mutate(data)
	}

	return map[string][]byte{eddystoneUUID: data}
}

func TestParseEddystoneTLM(t *testing.T) {
	got, ok := ParseEddystoneTLM(tlmFrame(nil))
	if !ok {
		t.Fatal("not recognised")
	}

	// 0x0bb8 is 3000 mV; 0x1880 is 24.5 degrees; 0x100 = 256 frames; 0x3e8 = 1000 tenths of a second.
	if got.BatteryMillivolts != 3000 || !got.HasTemperature || got.Temperature != 24.5 ||
		got.AdvertisingCount != 256 || got.Uptime != 100*time.Second {
		t.Errorf("got %+v", got)
	}
}

func TestParseEddystoneTLM_NegativeTemperatureAndNoThermometer(t *testing.T) {
	got, _ := ParseEddystoneTLM(tlmFrame(func(d []byte) { d[4], d[5] = 0xfe, 0x80 })) // -1.5
	if !got.HasTemperature || got.Temperature != -1.5 {
		t.Errorf("negative: %+v", got)
	}

	got, ok := ParseEddystoneTLM(tlmFrame(func(d []byte) { d[4], d[5] = 0x80, 0x00 }))
	if !ok || got.HasTemperature {
		t.Errorf("no thermometer: %+v %v", got, ok)
	}
}

func TestParseEddystoneTLM_RefusesWhatIsNotOne(t *testing.T) {
	for name, data := range map[string]map[string][]byte{
		"nothing":      nil,
		"another UUID": {"0000feab-0000-1000-8000-00805f9b34fb": make([]byte, 14)},
		"a UID frame":  tlmFrame(func(d []byte) { d[0] = 0x00 }),
		"encrypted":    tlmFrame(func(d []byte) { d[1] = 0x01 }),
		"short":        {eddystoneUUID: {0x20, 0x00, 0x0b}},
		"too long":     {eddystoneUUID: append(tlmFrame(nil)[eddystoneUUID], 0)},
	} {
		if _, ok := ParseEddystoneTLM(data); ok {
			t.Errorf("%s was accepted", name)
		}
	}
}
