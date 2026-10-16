package device

import (
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

// PowerMonitor is the generic, producer-facing representation of a device's electrical measurements, such
// as those of a smart plug: one entity per device, carrying whichever of the readings that device has.
//
// Devices differ in what they measure, so each reading has a Has flag that says whether the device has
// reported it. A reading whose flag is false is zero and means nothing; that isn't the same as measuring
// zero.
//
// The units are fixed, whatever the device uses: watts, volts, amps and kilowatt-hours.
type PowerMonitor struct {
	Name string

	// HasData is true once any reading has been reported.
	HasData bool

	// HasPower and Power are the power being drawn right now, in watts.
	HasPower bool
	Power    float64

	// HasVoltage and Voltage are the supply voltage, in volts.
	HasVoltage bool
	Voltage    float64

	// HasCurrent and Current are the current being drawn, in amps.
	HasCurrent bool
	Current    float64

	// HasEnergy and Energy are the energy used, in kilowatt-hours. What it counts from (the device's
	// lifetime, or since midnight) is up to the device, so don't compare it across devices.
	HasEnergy bool
	Energy    float64

	// LastUpdateTime is when any reading was last reported.
	LastUpdateTime time.Time

	// Connectivity is whether the device itself is currently reachable. The readings keep their last known
	// values while the device is offline, so consumers should check this before trusting them. OfflineSince
	// and OnlineSince are the same as on environment.Thermostat.
	Connectivity environment.Connectivity
	OfflineSince time.Time
	OnlineSince  time.Time
}

// IPowerMonitor is implemented by the stub a plugin hands out for a registered power monitor.
type IPowerMonitor interface {
	GetPowerMonitorData() PowerMonitor
}

// IPowerMonitorType is the entity type to register a power monitor under.
var IPowerMonitorType = reflect.TypeFor[IPowerMonitor]()
