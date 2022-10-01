package environemnt

import (
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
	Temperature    float32
	HasHumidity    bool
	Humidity       float32
	LastUpdateTime time.Time
}

var emptySensorData = SensorData{}

func (sd SensorData) IsEmpty() bool {
	return sd == emptySensorData
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
