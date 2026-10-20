package entity

// Ref names an entity before it exists. An assembly is configured before the plugins start, and an entity gets its id when its plugin
// registers it, so what the assembly holds is a handle (what a plugin's config gave back). A handle that implements Ref can say which
// entity it is once the entity is registered.
type Ref interface {
	// RegisteredId is the id of the entity, or false while it is not registered. Safe to call from any goroutine. (It is not called
	// EntityId because a handle may already have a method of that name that says something else.)
	RegisteredId() (string, bool)
}
