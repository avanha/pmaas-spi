package environment

import (
	"pmaas.io/spi/events"
)

type TemperatureChangeEvent struct {
	events.EntityEvent
	NewValue float32
	OldValue float32
}

type HumidityChangeEvent struct {
	events.EntityEvent
	NewValue float32
	OldValue float32
}
