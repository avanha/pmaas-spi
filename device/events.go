package device

import "github.com/avanha/pmaas-spi/events"

// DoorPositionChangedEvent is broadcast when a door sensor's position changes to a different known
// position. The very first report, which establishes the position rather than changing it, doesn't
// produce one - subscribe to events.EntityStateChangedEvent for that.
type DoorPositionChangedEvent struct {
	events.EntityEvent
	NewPosition DoorPosition
	OldPosition DoorPosition
}

// SwitchStateChangedEvent is broadcast when a switch's state changes to a different known state. The very
// first report, which establishes the state rather than changing it, doesn't produce one - subscribe to
// events.EntityStateChangedEvent for that.
type SwitchStateChangedEvent struct {
	events.EntityEvent
	NewState SwitchState
	OldState SwitchState
}
