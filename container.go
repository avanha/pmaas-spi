package spi

import (
	"fmt"
	"io/fs"
	"net/http"
	"reflect"

	"pmaas.io/spi/events"
)

type RenderListOptions struct {
	Title  string
	Header any
}

// IPMAASContainer is an interface for plugins to interact with the PMAAS server.
type IPMAASContainer interface {
	AddRoute(path string, handlerFunc http.HandlerFunc)
	BroadcastEvent(entityEventId string, event any) error
	RenderList(w http.ResponseWriter, r *http.Request, options RenderListOptions, items []interface{})
	GetTemplate(templateInfo *TemplateInfo) (CompiledTemplate, error)
	GetEntityRenderer(entityType reflect.Type) (EntityRenderer, error)
	RegisterEntityRenderer(entityType reflect.Type, renderFactory EntityRendererFactory)
	EnableStaticContent(staticContentDir string)

	// ProvideContentFS Registers an io/fs.FS instance that the server can use to read plugin resources such as
	// templates or static file content for serving over HTTP.
	ProvideContentFS(fs fs.FS, prefix string)

	// RegisterEntity Registers an entity with server.  This gives it a unique name, which can later be looked up for
	// further interaction with the entity.
	RegisterEntity(uniqueData string, entityType reflect.Type, name string,
		invocationHandlerFunc EntityInvocationHandlerFunc) (string, error)

	// DeregisterEntity Removes an entity previously registered with the server.  Pass the ID returned from the
	// previous call to RegisterEntity.
	DeregisterEntity(id string) error

	// AssertEntityType Verifies that an entity with the given ID exists and is of the passed type
	AssertEntityType(pmaasEntityId string, entityType reflect.Type) error

	// InvokeOnEntity Invokes the supplied function on the plugin-runner goroutine of the plugin that owns the specified
	//entity, supplying the entity.
	InvokeOnEntity(id string, function func(entity any)) error

	// RegisterEventReceiver Registers a receiver for events.  If successful, returns an integer handle that can be used
	// to deregister the handler in the future.
	RegisterEventReceiver(predicate events.EventPredicate, receiver events.EventReceiver) (int, error)

	// DeregisterEventReceiver Removes a previously registered event receiver
	DeregisterEventReceiver(receiverHandle int) error

	// EnqueueOnPluginGoRoutine Enqueues the passed function for execution on the plugin's main GoRoutine.
	// Returns an error if the function cannot be enqueued.  This method returns as soon as the function
	// is enqueued, it does not wait for execution.  If you need the results of the execution, you'll need to
	// orchestrate that in the function.  For example, the function can send the result back via a channel.
	// Warning: Calling this function when already executing on the plugin's main GoRoutine will result in deadlock.
	EnqueueOnPluginGoRoutine(f func()) error

	// EnqueueOnServerGoRoutine Enqueues the passed functions for execution on the server's main GoRoutine.
	// The server's main GoRoutine is the one used to call PMAAS.Run().
	// Use this to execute callbacks registered during the server configuration phase.
	EnqueueOnServerGoRoutine(invocations []func()) error
}

func ExecValueFunctionOnPluginGoRoutine[R any](container IPMAASContainer, f func() R, defaultValueFn func() R) (R, error) {
	resultCh := make(chan R)
	err := container.EnqueueOnPluginGoRoutine(func() {
		resultCh <- f()
		close(resultCh)
	})

	if err != nil {
		close(resultCh)

		return defaultValueFn(),
			fmt.Errorf("unable to enqueue value function execution on Plugin goroutine: %w", err)
	}

	result := <-resultCh

	return result, nil
}
