package common

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/avanha/pmaas-spi"
	"github.com/avanha/pmaas-spi/entity"
	"github.com/avanha/pmaas-spi/events"
)

// ErrStubNotAvailable is what a call through a StubHolder fails with while there is no stub to make it: the
// plugin hasn't registered the entity yet, or has deregistered it.
var ErrStubNotAvailable = errors.New("the entity's stub is not available")

// StubAvailableEvent is how an assembly is told that an entity's stub is available: the entity, and its stub.
type StubAvailableEvent[T any] struct {
	events.EntityEvent
	EntityStub T
}

// StubHolder is the assembly's side of an entity a plugin registers later. The assembly configures the plugin
// before the server runs, when there is no entity yet, and so no stub. A plugin that wants to let the assembly use
// its entities holds one StubHolder for each, in its configuration; the assembly's objects for the entity forward
// their calls through it, so that the assembly never has to find a stub, or wait for one.
//
// The plugin side calls Provide when it registers the entity, and Withdraw when it deregisters it. The assembly
// side calls Stub, and adds listeners. All of it is safe to use from any goroutine.
type StubHolder[T any] struct {
	mu        sync.Mutex
	stub      T
	available bool
	entityId  string
	listeners []func(event StubAvailableEvent[T])
}

var _ entity.Ref = (*StubHolder[any])(nil)

// RegisteredId is the id the server gave the entity, or false while there is no stub: it makes a StubHolder, and so any handle that
// has one, an entity.Ref, which is how an assembly names the entity to the groups plugin.
func (h *StubHolder[T]) RegisteredId() (string, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.entityId, h.available && h.entityId != ""
}

// AddOnStubAvailableListener adds a function to be called when the stub becomes available, on the server's
// goroutine, the one that called Run. It's called again if the stub is provided again. A listener added after the
// stub was provided is not called for it; Stub gives it that.
func (h *StubHolder[T]) AddOnStubAvailableListener(listener func(event StubAvailableEvent[T])) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.listeners = append(h.listeners, listener)
}

// Stub is the entity's stub, or an error wrapping ErrStubNotAvailable if there isn't one now.
func (h *StubHolder[T]) Stub() (T, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.available {
		var zero T
		return zero, ErrStubNotAvailable
	}

	return h.stub, nil
}

// Provide is for the plugin. It keeps event.EntityStub, so that Stub returns it from now on, and then hands the
// listeners to the server's goroutine, to be called with the event. They aren't called on the calling goroutine,
// which is the plugin's: they'll want to use the stub, which has to wait for it.
func (h *StubHolder[T]) Provide(container spi.IPMAASContainer, event StubAvailableEvent[T]) error {
	h.mu.Lock()
	h.stub = event.EntityStub
	h.available = true
	h.entityId = event.Id
	listeners := slices.Clone(h.listeners)
	h.mu.Unlock()

	if len(listeners) == 0 {
		return nil
	}

	invocations := make([]func(), len(listeners))
	for i, listener := range listeners {
		invocations[i] = func() { listener(event) }
	}

	if err := container.EnqueueOnServerGoRoutine(invocations); err != nil {
		return fmt.Errorf("unable to hand the stub of %s to the assembly's listeners: %w", event.Name, err)
	}

	return nil
}

// Withdraw is for the plugin. It takes the stub back, before the stub is closed: after it, Stub returns an error
// instead of a stub that is no longer any use.
func (h *StubHolder[T]) Withdraw() {
	h.mu.Lock()
	defer h.mu.Unlock()

	var zero T
	h.stub = zero
	h.available = false
	h.entityId = ""
}
