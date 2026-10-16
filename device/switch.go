package device

import (
	"errors"
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

// SwitchState is whether a switch (a relay, a smart plug) is currently on. The zero value,
// SwitchStateUnknown, deliberately isn't either real state, so a Switch that has never been told anything
// can't be mistaken for one that's off.
type SwitchState int

const (
	SwitchStateUnknown SwitchState = iota
	SwitchStateOn
	SwitchStateOff
)

func (s SwitchState) String() string {
	switch s {
	case SwitchStateOn:
		return "On"
	case SwitchStateOff:
		return "Off"
	default:
		return "Unknown"
	}
}

// ErrInvalidSwitchState is returned by ISwitch.SetState for a state a switch can't be put in: only
// SwitchStateOn and SwitchStateOff can be asked for. SwitchStateUnknown is only ever something a switch
// reports.
var ErrInvalidSwitchState = errors.New("a switch can only be set to on or off")

// IsSettable reports whether a switch can be asked to be in s: it's On or Off, and not Unknown.
func (s SwitchState) IsSettable() bool {
	return s == SwitchStateOn || s == SwitchStateOff
}

// Switch is the generic, producer-facing representation of something that can be turned on and off.
type Switch struct {
	Name string

	// HasData is true once the source has reported a state at all. Until then State is SwitchStateUnknown
	// and the time fields are zero.
	HasData bool

	// State is what the device last reported, not what was last asked of it: a command that the device
	// ignored or couldn't carry out leaves State as it was.
	State SwitchState

	// LastChangeTime is when State last changed to its current value, as observed by the source - not
	// merely when it was last reported (that's LastUpdateTime). Zero until a change has been observed: the
	// first report establishes the state but isn't known to be a change.
	LastChangeTime time.Time

	// LastUpdateTime is when the source last reported a state, changed or not.
	LastUpdateTime time.Time

	// Connectivity is whether the device itself is currently reachable. State keeps its last known value
	// while the device is offline, so consumers should check this before trusting it. OfflineSince and
	// OnlineSince are the same as on environment.Thermostat.
	Connectivity environment.Connectivity
	OfflineSince time.Time
	OnlineSince  time.Time
}

// ISwitch is implemented by the stub a plugin hands out for a registered switch.
type ISwitch interface {
	GetSwitchData() Switch

	// SetState asks the device to turn on (SwitchStateOn) or off (SwitchStateOff). It returns once the request
	// has been accepted for delivery, not once the device has acted on it: the new state arrives the way any
	// other does, as the device reports it, and shows up in GetSwitchData and as a SwitchStateChangedEvent.
	//
	// It returns an error if the request can't be delivered, for instance because the device is offline, and
	// ErrInvalidSwitchState (wrapped or not) for any state other than On or Off. Asking for the state the
	// switch is already in is allowed, and does nothing the device would notice.
	SetState(state SwitchState) error
}

// ISwitchType is the entity type to register a switch under.
var ISwitchType = reflect.TypeFor[ISwitch]()
