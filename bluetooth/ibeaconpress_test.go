package bluetooth_test

import (
	"testing"
	"time"

	"github.com/avanha/pmaas-spi/bluetooth"
	"github.com/avanha/pmaas-spi/bluetooth/bttest"
	"github.com/avanha/pmaas-spi/device"
)

// Frames from the BlueCharm BC011 Pro, captured from a proxy: a single click (slot 0) and a double (slot 1).
const (
	frameSingle = "0201061aff4c000215426c7565436861726d426561636f6e770efe1355c9"
	frameDouble = "0201061aff4c0002157777772e6b6b6d636e2e636f6d00000600010001c9"
)

func newPresses() *bluetooth.IBeaconPresses {
	p := bluetooth.NewIBeaconPresses("couch")
	p.Single(bluetooth.MatchIBeacon("426c7565-4368-6172-6d42-6561636f6e77"))
	p.Double(bluetooth.MatchAnyIBeacon().WithMajor(1))

	return p
}

func TestPressHearer_AnnouncesTheFirstFrameOfABurstOnly(t *testing.T) {
	h := newPresses().NewHearer()
	double := func() *bluetooth.AdvertisementEvent { return bttest.Advert(t, "DD:88:00:00:3D:36", frameDouble) }
	now := bttest.Start

	if got := h.Hear(double(), now); len(got) != 1 || got[0] != device.PressDouble {
		t.Fatalf("first frame: %v", got)
	}

	if got := h.Hear(double(), now.Add(2*time.Second)); got != nil {
		t.Errorf("the rest of the burst was announced: %v", got)
	}

	// A quiet spell longer than the burst and its margin: a new press. The frame at 2s kept the burst alive until 5.5s.
	if got := h.Hear(double(), now.Add(2*time.Second+4*time.Second)); len(got) != 1 {
		t.Errorf("a new press was not announced: %v", got)
	}
}

func TestPressHearer_TellsPressesApartAndIgnoresOtherFrames(t *testing.T) {
	h := newPresses().NewHearer()

	if got := h.Hear(bttest.Advert(t, "DD:88:00:00:3D:36", frameSingle), bttest.Start); len(got) != 1 || got[0] != device.PressSingle {
		t.Errorf("single: %v", got)
	}

	if got := h.Hear(bttest.Advert(t, "DD:88:00:00:3D:36", "020106"), bttest.Start); got != nil {
		t.Errorf("a frame with no iBeacon: %v", got)
	}
}

func TestIBeaconPresses_RefusesMistakes(t *testing.T) {
	mustPanic := func(name string, f func()) {
		t.Helper()

		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()

		f()
	}

	mustPanic("empty frame", func() { bluetooth.NewIBeaconPresses("x").Single(bluetooth.MatchAnyIBeacon()) })
	mustPanic("same frame twice", func() {
		p := bluetooth.NewIBeaconPresses("x")
		p.Single(bluetooth.MatchAnyIBeacon().WithMajor(1))
		p.Double(bluetooth.MatchAnyIBeacon().WithMajor(1))
	})
	mustPanic("burst", func() { bluetooth.NewIBeaconPresses("x").Burst(0) })

	if bluetooth.NewIBeaconPresses("x").Validate() == nil {
		t.Error("a button with no presses was valid")
	}
}
