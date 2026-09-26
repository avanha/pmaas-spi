package environment

import (
	"reflect"
	"time"
)

type RSSIData struct {
	RSSI           int
	LastUpdateTime time.Time
}

var emptyRSSIData = RSSIData{}

func (rd RSSIData) IsEmpty() bool {
	return rd == emptyRSSIData
}

type BatteryData struct {
	Level          int
	LastUpdateTime time.Time
}

var emptyBatteryData = BatteryData{}

func (bd BatteryData) IsEmpty() bool {
	return bd == emptyBatteryData
}

type SensorData struct {
	Temperature    float32 `track:"always"`
	HasHumidity    bool
	Humidity       float32   `track:"always,nullable"`
	LastUpdateTime time.Time `track:"always"`
}

var emptySensorData = SensorData{}

var SensorDataType = reflect.TypeOf((*SensorData)(nil)).Elem()

func (sd SensorData) IsEmpty() bool {
	return sd == emptySensorData
}

func SensorDataToInsertArgs(anyData *any) ([]any, error) {
	sd := (*anyData).(SensorData)
	var humidity any = nil

	if sd.HasHumidity {
		humidity = sd.Humidity
	}

	return []any{sd.Temperature, humidity, sd.LastUpdateTime}, nil
}

type WirelessThermometer struct {
	Name        string
	RSSIData    RSSIData
	BatteryData BatteryData
	SensorData  SensorData
}

type IWirelessThermometer interface {
	GetWirelessThermometerData() WirelessThermometer
}

// Thermostat is the generic, producer-facing representation of a thermostat-like device: a thermometer
// and hygrometer (via the embedded SensorData) plus the HVAC-specific attributes that distinguish a
// thermostat from a plain sensor. Any plugin whose device fits this shape can implement IThermostat to
// advertise it, the same way WirelessThermometer/IWirelessThermometer works for wireless sensors.
type Thermostat struct {
	Name       string
	SensorData SensorData

	// HvacStatus is what the system is actually doing right now: "OFF", "HEATING", "COOLING".
	HvacStatus string
	// Mode is the configured mode, independent of HvacStatus: "HEAT", "COOL", "HEATCOOL", "OFF". A
	// HEATCOOL-mode thermostat has both HeatSetpoint and CoolSetpoint meaningful at once, regardless of
	// whether HvacStatus currently shows it actively running either one.
	Mode    string
	EcoMode string

	HeatSetpoint float32
	CoolSetpoint float32

	// LastUpdateTime is the most recent update time across all of this thermostat's fields, for
	// consumers that just want a single "how fresh is this" value rather than per-field ones.
	LastUpdateTime time.Time
}

type IThermostat interface {
	GetThermostatData() Thermostat
}
