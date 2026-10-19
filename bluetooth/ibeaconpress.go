package bluetooth

import (
	"fmt"
	"time"

	"github.com/avanha/pmaas-spi/device"
)

// DefaultBurst is how long a beacon is assumed to go on advertising after a press: the BlueCharm's "Trigger Adv Duration"
// that most people set is a few seconds.
const DefaultBurst = 3 * time.Second

// burstMargin is how much longer than the burst a quiet spell has to be for the next frame to be a new press and not the
// rest of the last.
const burstMargin = 500 * time.Millisecond

// presses are all the ways a button can be pressed, in the order they are looked for.
var presses = []device.ButtonPress{device.PressSingle, device.PressDouble, device.PressTriple, device.PressLong}

// IBeaconPresses is which iBeacon frame is which press, for a button that is a beacon advertising a frame for a few seconds after
// it is pressed, a different one for each way. iBeacon is a standard and which frame is which press is the owner's choice, so this
// is not device knowledge: a builder for any such button uses it.
type IBeaconPresses struct {
	name   string
	burst  time.Duration
	frames map[device.ButtonPress]IBeaconMatch
}

func NewIBeaconPresses(name string) *IBeaconPresses {
	return &IBeaconPresses{name: name, burst: DefaultBurst, frames: make(map[device.ButtonPress]IBeaconMatch)}
}

// Single is the frame advertised after a single click. Each press needs a frame of its own. It panics if another press has the
// same frame, like regexp.MustCompile does for a pattern that can't be.
func (p *IBeaconPresses) Single(frame IBeaconMatch) { p.set(device.PressSingle, frame) }
func (p *IBeaconPresses) Double(frame IBeaconMatch) { p.set(device.PressDouble, frame) }
func (p *IBeaconPresses) Triple(frame IBeaconMatch) { p.set(device.PressTriple, frame) }
func (p *IBeaconPresses) Long(frame IBeaconMatch)   { p.set(device.PressLong, frame) }

// Burst is how long the beacon goes on advertising after a press (the BlueCharm's "Trigger Adv Duration"). Frames of the same
// press that come sooner than that, and a little longer, are the rest of the press that was already announced, not another press,
// so two presses of the same kind closer together than this are one. It panics if it is not positive.
func (p *IBeaconPresses) Burst(burst time.Duration) {
	if burst <= 0 {
		panic(fmt.Sprintf("bluetooth: button %q: the burst has to be longer than nothing, not %v", p.name, burst))
	}

	p.burst = burst
}

func (p *IBeaconPresses) set(press device.ButtonPress, frame IBeaconMatch) {
	if frame.IsEmpty() {
		panic(fmt.Sprintf("bluetooth: button %q: the %s press needs a frame that says something: use MatchIBeacon or MatchAnyIBeacon with WithMajor", p.name, press))
	}

	for other, existing := range p.frames {
		if other != press && existing.String() == frame.String() {
			panic(fmt.Sprintf("bluetooth: button %q: the %s and %s presses can't both be %s", p.name, other, press, frame))
		}
	}

	p.frames[press] = frame
}

// Validate is an error if no press has a frame.
func (p *IBeaconPresses) Validate() error {
	if len(p.frames) == 0 {
		return fmt.Errorf("the button %q needs at least one press: say which frame is Single, Double, Triple or Long", p.name)
	}

	return nil
}

// Which is the press a frame says, if it says one. A press whose match is more specific is looked for first, so that a UUID
// shared by two presses that differ in the minor is told apart.
func (p *IBeaconPresses) Which(frame IBeacon) (device.ButtonPress, bool) {
	var found device.ButtonPress
	specificity := -1

	for _, press := range presses {
		match, ok := p.frames[press]
		if !ok || !match.Matches(frame) {
			continue
		}

		if score := match.Specificity(); score > specificity {
			found, specificity = press, score
		}
	}

	return found, specificity >= 0
}

// NewHearer is the state for one button: it tells a new press from the rest of the burst of one already announced.
func (p *IBeaconPresses) NewHearer() *PressHearer {
	return &PressHearer{presses: p, last: make(map[device.ButtonPress]time.Time)}
}

// PressHearer finds the presses in a button's advertisements.
type PressHearer struct {
	presses *IBeaconPresses
	last    map[device.ButtonPress]time.Time
}

// Hear is the press the advertisement starts, if it is the first frame of one.
func (h *PressHearer) Hear(a *AdvertisementEvent, now time.Time) []device.ButtonPress {
	frame, ok := ParseIBeacon(a.ManufacturerData)
	if !ok {
		return nil
	}

	press, ok := h.presses.Which(frame)
	if !ok {
		return nil
	}

	last, heard := h.last[press]
	h.last[press] = now

	if heard && now.Sub(last) < h.presses.burst+burstMargin {
		return nil
	}

	return []device.ButtonPress{press}
}
