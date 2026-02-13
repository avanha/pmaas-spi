package predicates

import (
	"pmaas.io/spi/events"
	"reflect"
)

func And(predicate1 events.EventPredicate, predicate2 events.EventPredicate) events.EventPredicate {
	return func(eventInfo *events.EventInfo) bool {
		return predicate1(eventInfo) && predicate2(eventInfo)
	}
}

func Or(predicate1 events.EventPredicate, predicate2 events.EventPredicate) events.EventPredicate {
	return func(eventInfo *events.EventInfo) bool {
		return predicate1(eventInfo) || predicate2(eventInfo)
	}
}

func EventTypeEquals(eventType reflect.Type) events.EventPredicate {
	return func(eventInfo *events.EventInfo) bool {
		return reflect.TypeOf(eventInfo.Event) == eventType
	}
}
