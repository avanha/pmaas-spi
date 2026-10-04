// Package device holds the producer-facing representations of home devices that aren't
// environmental sensors (see the environment package for those): a plugin that fits one of these
// shapes implements the matching interface and registers itself under its reflect.Type, the same way
// environment.IThermostat and environment.IWirelessThermometer work.
package device

import (
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

// DoorPosition is where a door (or garage door) currently is. The zero value, DoorPositionUnknown,
// deliberately isn't either real position, so a DoorSensor that has never been told anything can't
// be mistaken for a closed door.
type DoorPosition int

const (
	DoorPositionUnknown DoorPosition = iota
	DoorPositionOpen
	DoorPositionClosed
)

func (p DoorPosition) String() string {
	switch p {
	case DoorPositionOpen:
		return "Open"
	case DoorPositionClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}

// DoorSensor is the generic, producer-facing representation of a door position detector.
type DoorSensor struct {
	Name string

	// HasData is true once the source has reported a position at all. Until then Position is
	// DoorPositionUnknown and the time fields are zero.
	HasData  bool
	Position DoorPosition

	// LastChangeTime is when Position last changed to its current value, as observed by the source
	// - not merely when it was last reported (that's LastUpdateTime). Zero until a change has been
	// observed: the first report establishes the position but isn't known to be a change.
	LastChangeTime time.Time

	// LastUpdateTime is when the source last reported a position, changed or not.
	LastUpdateTime time.Time

	// Connectivity is whether the detector itself is currently reachable. Position keeps its
	// last known value while the detector is offline, so consumers should check this before
	// trusting it. OfflineSince and OnlineSince are the same as on environment.Thermostat.
	Connectivity environment.Connectivity
	OfflineSince time.Time
	OnlineSince  time.Time
}

// IDoorSensor is implemented by the stub a plugin hands out for a registered door position detector.
type IDoorSensor interface {
	GetDoorSensorData() DoorSensor
}

// IDoorSensorType is the entity type to register a door position detector under.
var IDoorSensorType = reflect.TypeFor[IDoorSensor]()
