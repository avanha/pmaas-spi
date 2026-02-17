package entity

import "reflect"

type RegisteredEntityInfo struct {
	Id            string
	EntityType    reflect.Type
	Name          string
	StubFactoryFn func() (any, error)
}
