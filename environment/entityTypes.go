package environemnt

import (
	"time"
)

type RSSIData struct {
	RSSI	int
	LastUpdateTime time.Time
}

type BatteryData struct {
	Level int
	LastUpdateTime time.Time
}

type SensorData struct {
	Temperature float32
	HasHumidity bool
	Humidity float32
	LastUpdateTime time.Time
}

type WirelessThermometer struct {
	Name string
	RSSIData RSSIData
	BatteryData BatteryData
	SensorData SensorData
}