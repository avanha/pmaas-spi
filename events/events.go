package events

import "reflect"

type EntityRegisteredEvent struct {
	Id         string
	EntityType reflect.Type
}

type EventInfo struct {
	EventSourceType reflect.Type
	Event           any
}

type EventPredicate func(eventInfo *EventInfo) bool

type EventReceiver func(eventInfo *EventInfo) error
