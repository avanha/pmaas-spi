package events

import (
	"reflect"
)

type EntityEvent struct {
	Id         string
	EntityType reflect.Type
	Name       string
}

type EntityRegisteredEvent struct {
	EntityEvent
	StubFactoryFn func() (any, error)
}

type EntityDeregisteredEvent struct {
	EntityEvent
}

type EntityNameChangedEvent struct {
	EntityEvent
	NewName string
	OldName string
}

type EntityStateChangedEvent struct {
	EntityEvent
	NewState any
}

type EventInfo struct {
	SourceEntityId   string
	SourcePluginType reflect.Type
	Event            any
}

type EventPredicate func(eventInfo *EventInfo) bool

type EventReceiver func(eventInfo *EventInfo) error
