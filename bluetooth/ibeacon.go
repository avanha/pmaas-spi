package bluetooth

import (
	"fmt"
)

// appleCompanyID is the company ID an iBeacon frame is advertised under, in manufacturer data.
const appleCompanyID = 0x004c

// iBeaconLength is the length of the data after the company ID in an iBeacon frame: a type of 0x02, a length of 0x15 (21), the
// UUID (16), the major (2), the minor (2) and the measured power (1).
const iBeaconLength = 23

// IBeacon is what an iBeacon frame says: a UUID, a major and a minor number, which a beacon's owner chooses to say whose it
// is and what it is for, and the measured power, which is the signal strength at one metre.
type IBeacon struct {
	// UUID is 128-bit and lower case, in the form Fields uses: "426c7565-4368-6172-6d42-6561636f6e73".
	UUID          string
	Major, Minor  uint16
	MeasuredPower int8
}

// ParseIBeacon finds the iBeacon frame in an advertisement's manufacturer data (see Fields), if there is one.
func ParseIBeacon(manufacturerData map[uint16][]byte) (IBeacon, bool) {
	data, ok := manufacturerData[appleCompanyID]
	if !ok || len(data) != iBeaconLength || data[0] != 0x02 || data[1] != 0x15 {
		return IBeacon{}, false
	}

	return IBeacon{
		UUID:          fmt.Sprintf("%x-%x-%x-%x-%x", data[2:6], data[6:8], data[8:10], data[10:12], data[12:18]),
		Major:         uint16(data[18])<<8 | uint16(data[19]),
		Minor:         uint16(data[20])<<8 | uint16(data[21]),
		MeasuredPower: int8(data[22]),
	}, true
}

// IBeaconMatch says which iBeacon frames it is: those with the UUID, if it is set, and the major and the minor, if they are
// set. A match with none of them set would be every iBeacon, and is not a useful thing to ask for.
type IBeaconMatch struct {
	UUID         string
	Major, Minor *uint16
}

// MatchAnyIBeacon is a match of iBeacon frames of any UUID, to be narrowed with WithMajor and WithMinor. It is for a beacon
// whose frames are told apart by the numbers rather than the UUID, as the slots of a BlueCharm can be.
func MatchAnyIBeacon() IBeaconMatch {
	return IBeaconMatch{}
}

// IsEmpty is whether nothing has been said about which frames: it would match every one.
func (m IBeaconMatch) IsEmpty() bool {
	return m.UUID == "" && m.Major == nil && m.Minor == nil
}

// Specificity is how many of the UUID, the major and the minor the match requires, for choosing among matches that all
// fit a frame: the more specific one is the better fit.
func (m IBeaconMatch) Specificity() int {
	score := 0

	if m.UUID != "" {
		score++
	}

	if m.Major != nil {
		score++
	}

	if m.Minor != nil {
		score++
	}

	return score
}

// MatchIBeacon matches the frames with the UUID, which may be in any form NormalizeUUID understands. It panics if it is not a
// UUID, like regexp.MustCompile does for a pattern that can't be: it is written in the program.
func MatchIBeacon(uuid string) IBeaconMatch {
	normal, ok := NormalizeUUID(uuid)
	if !ok {
		panic(fmt.Sprintf("bluetooth.MatchIBeacon: %q is not a UUID", uuid))
	}

	return IBeaconMatch{UUID: normal}
}

// WithMajor also requires the major number to be this.
func (m IBeaconMatch) WithMajor(major uint16) IBeaconMatch {
	m.Major = &major

	return m
}

// WithMinor also requires the minor number to be this.
func (m IBeaconMatch) WithMinor(minor uint16) IBeaconMatch {
	m.Minor = &minor

	return m
}

// Matches is whether the frame is one this matches.
func (m IBeaconMatch) Matches(frame IBeacon) bool {
	return (m.UUID == "" || frame.UUID == m.UUID) &&
		(m.Major == nil || *m.Major == frame.Major) && (m.Minor == nil || *m.Minor == frame.Minor)
}

func (m IBeaconMatch) String() string {
	text := m.UUID
	if text == "" {
		text = "any UUID"
	}

	if m.Major != nil {
		text += fmt.Sprintf(" major %d", *m.Major)
	}

	if m.Minor != nil {
		text += fmt.Sprintf(" minor %d", *m.Minor)
	}

	return text
}
