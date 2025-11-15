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

func NewTrackableStub(container spi.IPMAASContainer, pmaasId string) Trackable {
	invoker := func(f func(trackable Trackable)) error {
		return container.InvokeOnEntity(pmaasId, func(entity any) {
			f(entity.(Trackable))
		})
	}
	return &trackableStub{
		pmaasEntityId: pmaasId,
		invokeFn:      invoker,
	}
}

type trackableStub struct {
	pmaasEntityId string
	invokeFn      func(func(trackable Trackable)) error
}

func (t *trackableStub) TrackingConfig() TrackingConfig {
	resultCh := make(chan TrackingConfig)
	err := t.invokeFn(func(instance Trackable) {
		defer func() { close(resultCh) }()
		resultCh <- instance.TrackingConfig()
	})

	if err != nil {
		close(resultCh)
		panic(err)
	}

	return <-resultCh
}

func (t *trackableStub) Data() any {
	resultCh := make(chan any)
	err := t.invokeFn(func(instance Trackable) {
		defer func() { close(resultCh) }()
		resultCh <- instance.Data()
	})

	if err != nil {
		close(resultCh)
		panic(err)
	}

	return <-resultCh
}
