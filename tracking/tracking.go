package tracking

import (
	"reflect"

	"pmaas.io/spi"
)

const TRACKING_MODE_POLL = 1
const TRACKING_MODE_PUSH = 2

type TrackingConfig struct {
	Name         string
	TrackingMode int
	Schema       string
}

type Trackable interface {
	TrackingConfig() TrackingConfig
	Data() any
}

var TrackableType = reflect.TypeOf((*Trackable)(nil)).Elem()

func NewTrackableStub(entityId string, container spi.IPMAASContainer) *TrackableStub {
	return &TrackableStub{
		entityId:  entityId,
		container: container,
	}
}

type TrackableStub struct {
	entityId  string
	container spi.IPMAASContainer
}

func (t *TrackableStub) TrackingConfig() TrackingConfig {
	var trackingConfig TrackingConfig
	err := t.container.InvokeOnEntity(t.entityId, func(entity any) {
		trackingConfig = entity.(Trackable).TrackingConfig()
	})

	if err != nil {
		panic(err)
	}

	return trackingConfig
}

func (t *TrackableStub) Data() any {
	var data any
	err := t.container.InvokeOnEntity(t.entityId, func(entity any) {
		data = entity.(Trackable).Data()
	})

	if err != nil {
		panic(err)
	}

	return data
}
