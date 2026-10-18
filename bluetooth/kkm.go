package bluetooth

import "fmt"

const (
	kkmSystemType      = 0x22
	kkmSystemMinLength = 11
)

// KKMSystem is what a KKM "System" frame says. KKM makes the beacons sold as BlueCharm (the BC011 Pro among them), and the frame
// is theirs rather than a standard: it carries who the beacon is and what its battery is at, and is advertised in the same service
// data as Eddystone's frames (0xFEAA). Its layout is in KKM's open source SDK (KBAdvPacketSystem): a frame type of 0x22, the model,
// the battery as a percentage, the beacon's MAC address and the firmware version, major and minor.
type KKMSystem struct {
	Model uint8

	// BatteryPercent is the beacon's own figure for its battery, which can be over 100 on a new cell.
	BatteryPercent uint8

	// Address is the beacon's address, "DD:88:00:00:3D:36".
	Address string

	FirmwareMajor, FirmwareMinor uint8
}

// ParseKKMSystem finds the KKM System frame in an advertisement's service data (see Fields), if there is one.
func ParseKKMSystem(serviceData map[string][]byte) (KKMSystem, bool) {
	data, ok := serviceData[eddystoneUUID]
	if !ok || len(data) < kkmSystemMinLength || data[0] != kkmSystemType {
		return KKMSystem{}, false
	}

	return KKMSystem{
		Model:          data[1],
		BatteryPercent: data[2],
		Address:        fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", data[3], data[4], data[5], data[6], data[7], data[8]),
		FirmwareMajor:  data[9],
		FirmwareMinor:  data[10],
	}, true
}
