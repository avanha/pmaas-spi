package bluetooth

import "time"

// eddystoneUUID is the 16-bit service UUID 0xFEAA that Eddystone frames are advertised under, in the form Fields uses.
const eddystoneUUID = "0000feaa-0000-1000-8000-00805f9b34fb"

const (
	eddystoneTLMType   = 0x20
	eddystoneTLMLength = 14
)

// EddystoneTLM is what an Eddystone-TLM frame says: the telemetry a beacon advertises about itself, in the frame beacons of
// the BlueCharm kind can be set to advertise alongside the ones that say who they are.
type EddystoneTLM struct {
	// BatteryMillivolts is the battery voltage, or 0 if the beacon doesn't say.
	BatteryMillivolts uint16

	// Temperature is the beacon's temperature in degrees Celsius, if it says (see HasTemperature).
	Temperature    float32
	HasTemperature bool

	// AdvertisingCount is how many frames the beacon has sent since it was switched on.
	AdvertisingCount uint32

	// Uptime is how long it has been on.
	Uptime time.Duration
}

// ParseEddystoneTLM finds the unencrypted Eddystone-TLM frame in an advertisement's service data (see Fields), if there is one.
func ParseEddystoneTLM(serviceData map[string][]byte) (EddystoneTLM, bool) {
	data, ok := serviceData[eddystoneUUID]
	if !ok || len(data) != eddystoneTLMLength || data[0] != eddystoneTLMType || data[1] != 0x00 {
		return EddystoneTLM{}, false
	}

	tlm := EddystoneTLM{
		BatteryMillivolts: uint16(data[2])<<8 | uint16(data[3]),
		AdvertisingCount:  uint32(data[6])<<24 | uint32(data[7])<<16 | uint32(data[8])<<8 | uint32(data[9]),
		Uptime:            time.Duration(uint32(data[10])<<24|uint32(data[11])<<16|uint32(data[12])<<8|uint32(data[13])) * 100 * time.Millisecond,
	}

	// The temperature is a signed 8.8 fixed point number, and 0x8000 is how a beacon says it has no thermometer.
	if raw := uint16(data[4])<<8 | uint16(data[5]); raw != 0x8000 {
		tlm.Temperature, tlm.HasTemperature = float32(int16(raw))/256, true
	}

	return tlm, true
}
