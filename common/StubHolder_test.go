package common_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/avanha/pmaas-spi"
	"github.com/avanha/pmaas-spi/common"
	"github.com/avanha/pmaas-spi/events"
)

type greeter interface{ Greet() string }

type greeterStub string

func (g greeterStub) Greet() string { return string(g) }

type fakeContainer struct {
	spi.IPMAASContainer

	batches [][]func()
	err     error
}

func (c *fakeContainer) EnqueueOnServerGoRoutine(invocations []func()) error {
	if c.err != nil {
		return c.err
	}

	c.batches = append(c.batches, invocations)

	return nil
}

func availableEvent(stub greeter) common.StubAvailableEvent[greeter] {
	return common.StubAvailableEvent[greeter]{EntityEvent: events.EntityEvent{Id: "id1", Name: "Greeter"}, EntityStub: stub}
}

func TestStubHolder_NoStubUntilProvided(t *testing.T) {
	var holder common.StubHolder[greeter]

	stub, err := holder.Stub()
	if !errors.Is(err, common.ErrStubNotAvailable) || stub != nil {
		t.Fatalf("got %v, %v", stub, err)
	}

	if err := holder.Provide(&fakeContainer{}, availableEvent(greeterStub("hello"))); err != nil {
		t.Fatal(err)
	}

	stub, err = holder.Stub()
	if err != nil || stub.Greet() != "hello" {
		t.Fatalf("got %v, %v", stub, err)
	}
}

func TestStubHolder_ListenersAreHandedToTheServerNotRunOnTheCaller(t *testing.T) {
	var holder common.StubHolder[greeter]
	container := &fakeContainer{}

	var heard []common.StubAvailableEvent[greeter]
	var stubWhenHeard []string

	holder.AddOnStubAvailableListener(func(event common.StubAvailableEvent[greeter]) {
		heard = append(heard, event)
		// By the time a listener runs, the holder can already be used, which is what lets the listener be the one
		// that does.
		stub, err := holder.Stub()
		if err != nil {
			t.Errorf("the holder had no stub in a listener: %v", err)
			return
		}

		stubWhenHeard = append(stubWhenHeard, stub.Greet())
	})

	if err := holder.Provide(container, availableEvent(greeterStub("hello"))); err != nil {
		t.Fatal(err)
	}

	if len(heard) != 0 || len(container.batches) != 1 || len(container.batches[0]) != 1 {
		t.Fatalf("heard %d, batches %v: the listener must be handed over, not run", len(heard), container.batches)
	}

	container.batches[0][0]()

	if len(heard) != 1 || heard[0].Id != "id1" || heard[0].Name != "Greeter" || stubWhenHeard[0] != "hello" {
		t.Fatalf("heard %+v, stub %v", heard, stubWhenHeard)
	}
}

func TestStubHolder_NobodyListeningMeansNothingForTheServerToRun(t *testing.T) {
	var holder common.StubHolder[greeter]
	container := &fakeContainer{}

	if err := holder.Provide(container, availableEvent(greeterStub("hello"))); err != nil {
		t.Fatal(err)
	}

	if len(container.batches) != 0 {
		t.Errorf("the server was asked to run %d batches with no listeners", len(container.batches))
	}
}

func TestStubHolder_TheStubIsKeptEvenIfTheServerCantRunTheListeners(t *testing.T) {
	var holder common.StubHolder[greeter]
	holder.AddOnStubAvailableListener(func(common.StubAvailableEvent[greeter]) {})
	container := &fakeContainer{err: errors.New("stopping")}

	err := holder.Provide(container, availableEvent(greeterStub("hello")))
	if err == nil {
		t.Fatal("no error although the listeners could not be handed over")
	}

	if stub, stubErr := holder.Stub(); stubErr != nil || stub.Greet() != "hello" {
		t.Errorf("got %v, %v", stub, stubErr)
	}
}

func TestStubHolder_WithdrawTakesTheStubBack(t *testing.T) {
	var holder common.StubHolder[greeter]
	_ = holder.Provide(&fakeContainer{}, availableEvent(greeterStub("hello")))

	holder.Withdraw()

	if stub, err := holder.Stub(); !errors.Is(err, common.ErrStubNotAvailable) || stub != nil {
		t.Errorf("got %v, %v", stub, err)
	}
}

func TestStubHolder_ProvidedAgainItReplacesTheStubAndTellsTheListenersAgain(t *testing.T) {
	var holder common.StubHolder[greeter]
	container := &fakeContainer{}
	calls := 0
	holder.AddOnStubAvailableListener(func(common.StubAvailableEvent[greeter]) { calls++ })

	_ = holder.Provide(container, availableEvent(greeterStub("first")))
	holder.Withdraw()
	_ = holder.Provide(container, availableEvent(greeterStub("second")))

	for _, batch := range container.batches {
		for _, invocation := range batch {
			invocation()
		}
	}

	if stub, _ := holder.Stub(); stub.Greet() != "second" || calls != 2 {
		t.Errorf("stub %v, calls %d", stub, calls)
	}
}

func TestStubHolder_ListenersAddedLaterAreNotCalledForAnEarlierStub(t *testing.T) {
	var holder common.StubHolder[greeter]
	container := &fakeContainer{}
	_ = holder.Provide(container, availableEvent(greeterStub("hello")))

	holder.AddOnStubAvailableListener(func(common.StubAvailableEvent[greeter]) { t.Error("called") })

	if len(container.batches) != 0 {
		t.Errorf("batches: %d", len(container.batches))
	}
}

func TestStubHolder_SafeToUseFromManyGoroutines(t *testing.T) {
	var holder common.StubHolder[greeter]
	container := &fakeContainer{}

	var wg sync.WaitGroup
	var mu sync.Mutex // the fake container isn't safe by itself

	for i := 0; i < 8; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			holder.AddOnStubAvailableListener(func(common.StubAvailableEvent[greeter]) {})
		}()
		go func() { defer wg.Done(); _, _ = holder.Stub() }()
		go func() {
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			_ = holder.Provide(container, availableEvent(greeterStub("x")))
			holder.Withdraw()
		}()
	}

	wg.Wait()
}
