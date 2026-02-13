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
