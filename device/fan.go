package device

import (
	"errors"
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

// FanDirection is which way a ceiling fan turns. The zero value, FanDirectionUnknown, is only ever something a fan reports.
type FanDirection int

const (
	FanDirectionUnknown FanDirection = iota
	FanDirectionForward
	FanDirectionReverse
)

func (d FanDirection) String() string {
	switch d {
	case FanDirectionForward:
		return "Forward"
	case FanDirectionReverse:
		return "Reverse"
	default:
		return "Unknown"
	}
}

// ErrInvalidFanSpeed is returned by IFan.SetSpeed for a speed the fan doesn't have.
var ErrInvalidFanSpeed = errors.New("the fan doesn't have that speed")

// ErrInvalidFanDirection is returned by IFan.SetDirection for anything but forward or reverse.
var ErrInvalidFanDirection = errors.New("a fan can only be set to forward or reverse")

// Fan is the generic, producer-facing representation of a fan: something that turns on and off, has a number of speeds, and may
// turn either way.
type Fan struct {
	Name string

	// HasData is true once the source has said anything about the fan.
	HasData bool

	// State is whether the fan is on. Speed is 1..MaxSpeed while it is, and the speed it will come on at when it is not.
	State     SwitchState
	Speed     int
	MaxSpeed  int
	Direction FanDirection

	// Assumed is that the source can't read the fan: it knows only what it last sent, so a change made some other way, with the
	// fan's own remote, is not seen.
	Assumed bool

	LastChangeTime time.Time
	LastUpdateTime time.Time

	// Connectivity is whether the device that controls the fan is reachable. See Switch.
	Connectivity environment.Connectivity
	OfflineSince time.Time
	OnlineSince  time.Time
}

// IFan is implemented by the stub a plugin hands out for a registered fan. It is also an ISwitch, for on and off, so a fan goes
// wherever a switch does (an automation, a button).
type IFan interface {
	ISwitch
	GetFanData() Fan

	// SetSpeed turns the fan on at the speed, 1..MaxSpeed. As with SetState, it returns once the request is accepted for delivery.
	SetSpeed(speed int) error

	// SetDirection asks for FanDirectionForward or FanDirectionReverse.
	SetDirection(direction FanDirection) error
}

// IFanType is the entity type to register a fan under.
var IFanType = reflect.TypeFor[IFan]()
