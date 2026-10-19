package device

import (
	"errors"
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

// ErrNotDimmable is returned by ILight for a brightness change on a light that can't be dimmed.
var ErrNotDimmable = errors.New("the light can't be dimmed")

// ErrInvalidBrightness is returned by ILight.SetBrightness for a level outside 1..100.
var ErrInvalidBrightness = errors.New("the brightness has to be from 1 to 100 percent")

// Light is the generic, producer-facing representation of a light that may be dimmed.
type Light struct {
	Name string

	// HasData is true once the source has said anything about the light.
	HasData bool

	State SwitchState

	// Dimmable is whether the brightness can be changed. Brightness is a percentage, from 1 to 100, and is only meaningful if it
	// is Dimmable and HasBrightness.
	Dimmable      bool
	HasBrightness bool
	Brightness    int

	// Assumed is that the source can't read the light, so State is what it last sent. BrightnessEstimated is that the brightness
	// is worked out from how long the dimmer was run, not read, so it drifts: it is as good as the timing.
	Assumed             bool
	BrightnessEstimated bool

	LastChangeTime time.Time
	LastUpdateTime time.Time

	// Connectivity is whether the device that controls the light is reachable. See Switch.
	Connectivity environment.Connectivity
	OfflineSince time.Time
	OnlineSince  time.Time
}

// ILight is implemented by the stub a plugin hands out for a registered light. It is also an ISwitch, for on and off, so a light
// goes wherever a switch does.
type ILight interface {
	ISwitch
	GetLightData() Light

	// SetBrightness asks for a brightness from 1 to 100 percent, and turns the light on. As with SetState, it returns once the
	// request is accepted for delivery: a dimmer that has to be run for a while is finished later.
	SetBrightness(percent int) error

	// StepBrightness makes the light brighter by a number of percent, or dimmer if it is negative, from where it is.
	StepBrightness(percent int) error
}

// ILightType is the entity type to register a light under.
var ILightType = reflect.TypeFor[ILight]()
