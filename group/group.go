// Package group is a way to organise entities: a tree of named groups, each with the entities that are in it directly. An assembly
// builds the tree when it configures the groups plugin. Each group is an entity as well, so that a plugin or an automation can ask
// what is in one.
package group

import "reflect"

// Group is a group as the groups plugin knows it.
type Group struct {
	// Name is the group's own name, such as "Master Bedroom".
	Name string

	// Path is the names from the root to the group, joined with "/": "Cougar/Master Bedroom". The root's is empty.
	Path string

	// Parent is the path of the group above, empty for the root and for a group directly under it.
	Parent string

	// Children are the paths of the groups directly under it, in order.
	Children []string

	// Entities are the ids of the entities directly in the group that exist, in order.
	Entities []string
}

// IGroup is a group as an entity. The stub is safe to use from any goroutine.
type IGroup interface {
	// GetGroupData is the group as it is now.
	GetGroupData() Group

	// Descendants are the ids of the entities in the group and in the groups below it, each once, direct members first.
	Descendants() []string
}

// IGroupType is the type groups are registered with.
var IGroupType = reflect.TypeFor[IGroup]()
