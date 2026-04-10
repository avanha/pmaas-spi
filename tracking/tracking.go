package tracking

import (
	"reflect"
	"time"
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
