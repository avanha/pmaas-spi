package tracking

import (
	"reflect"

	"pmaas.io/spi"
)

const ModePoll = 1
const ModePush = 2

type Config struct {
	Name         string
	TrackingMode int
	Schema       string
}

func (tc *Config) Clone() Config {
	return *tc
}

type Trackable interface {
	TrackingConfig() Config
	Data() any
}

var TrackableType = reflect.TypeOf((*Trackable)(nil)).Elem()

func CreateTrackableStub(container spi.IPMAASContainer, pmaasId string) Trackable {
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

func (t *trackableStub) TrackingConfig() Config {
	resultCh := make(chan Config)
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
