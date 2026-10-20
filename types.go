package spi

import (
	"io"
	"reflect"
)

type EntityInvocationHandlerFunc func(func(entity any)) error

type EntityStubFactoryFunc func() (any, error)

// EntityViewFunc makes the view of an entity: the value a page draws a card for it from. The plugin that registered the entity gives
// one for it (see IPMAASContainer.RegisterEntityView), and the server runs it on that plugin's goroutine, so it may read the plugin's
// state. The value's type is what finds its renderer (see RegisterEntityRenderer), the same one the plugin's own page uses.
type EntityViewFunc func() (EntityView, error)

// EntityView is what an entity looks like on a page.
type EntityView struct {
	// Key says which card this is. Entities whose views have the same key are one card: a device's card holds its children, which are
	// entities of their own, and a page that lists them all draws the card once. Empty means the view is the entity's alone.
	Key string

	// Category is what kind of thing the card is, for a page that sorts entities by kind: one of the group package's Category
	// constants, or a name of the plugin's own. A card is one kind even if it holds several (a device with a fan and a light is a
	// fan). Empty leaves it to the page, which goes by the entity's registered type.
	Category string

	// Covers are the entities whose own cards this one makes redundant: a plugin that shows another plugin's entity in a fuller card
	// (the environment plugin's thermometers) names the entity it shows, so a page that lists them all draws the fuller card and not
	// both.
	Covers []string

	// Value is what the renderer is given.
	Value any
}

// EntityViewResult is the answer to a request for the view of an entity (see IPMAASContainer.GetEntityViews).
type EntityViewResult struct {
	EntityId   string
	Name       string
	EntityType reflect.Type

	// Plugin is the short name of the plugin that owns the entity, and PluginPath where its page is.
	Plugin     string
	PluginPath string

	// View is the entity's view, or nil if its plugin gave none: all there is to show of it is the above.
	View *EntityView

	// Err is why the view could not be made, if it could not. The rest is filled in nevertheless.
	Err error
}

type EntityRenderFunc func(entity any) (string, error)

type StreamingEntityRenderFunc func(w io.Writer, entity any) error

type EntityRendererFactory func() (EntityRenderer, error)

type EntityRenderer struct {
	RenderFunc          EntityRenderFunc
	StreamingRenderFunc StreamingEntityRenderFunc
	Styles              []string
	Scripts             []string
}
