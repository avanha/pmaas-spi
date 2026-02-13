package tracking

import (
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi"
)

const ModePoll = 1
const ModePush = 2

type InsertArgFactoryFunc func(*any) ([]any, error)

type Schema struct {
	DataStructType     reflect.Type
	InsertArgFactoryFn InsertArgFactoryFunc
}

type Config struct {
	Name                string
	TrackingMode        int
	PollIntervalSeconds int
	Schema              Schema
}

type DataSample struct {
	LastUpdateTime time.Time
	Data           any
}

func (tc *Config) Clone() Config {
	return *tc
}

type Trackable interface {
	TrackingConfig() Config
	Data() DataSample
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

func (t *trackableStub) Data() DataSample {
	resultCh := make(chan DataSample)
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
