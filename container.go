package spi

import (
	"net/http"
	"pmaas.io/spi/events"
	"reflect"
)

type RenderListOptions struct {
	Title string
}

// IPMAASContainer is an interface for plugins to interact with the PMAAS server.
type IPMAASContainer interface {
	AddRoute(path string, handlerFunc http.HandlerFunc)
	BroadcastEvent(event any) error
	RenderList(w http.ResponseWriter, r *http.Request, options RenderListOptions, items []interface{})
	GetTemplate(templateInfo *TemplateInfo) (CompiledTemplate, error)
	GetEntityRenderer(entityType reflect.Type) (EntityRenderer, error)
	RegisterEntityRenderer(entityType reflect.Type, renderFactory EntityRendererFactory)
	EnableStaticContent(staticContentDir string)

	// RegisterEntity Registers an entity with server.  This gives it a unique name, which can later be looked up for
	// further interaction with the entity.
	RegisterEntity(uniqueData string, entityType reflect.Type, name string) (string, error)

	// DeregisterEntity Removes an entity previously registered with the server.  Pass the id returned from the
	// previous call to RegisterEntity.
	DeregisterEntity(id string) error

	// RegisterEventReceiver Registers a receiver for events.  If successful, returns an integer handle that can be used
	// to deregister the handler in the future.
	RegisterEventReceiver(predicate events.EventPredicate, receiver events.EventReceiver) (int, error)

	// DeregisterEventReceiver Removes a previously registered event receiver
	DeregisterEventReceiver(receiverHandle int) error
}
