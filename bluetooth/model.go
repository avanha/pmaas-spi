package bluetooth

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/avanha/pmaas-spi"
	"github.com/avanha/pmaas-spi/common"
	"github.com/avanha/pmaas-spi/device"
)

// Kind is what a device is, which decides the entity the bluetooth plugin registers for it.
type Kind int

const (
	KindThermometer Kind = iota + 1
	KindButton
)

func (k Kind) String() string {
	switch k {
	case KindThermometer:
		return "thermometer"
	case KindButton:
		return "button"
	default:
		return "unknown"
	}
}

// Environment is a temperature (Celsius) and, if the device has one, a humidity (percent).
type Environment struct {
	Temperature float32
	Humidity    float32
	HasHumidity bool
}

// Result is what a Device made of an advertisement. A nil field says nothing.
type Result struct {
	Battery    *int // percent
	Millivolts *int // the battery's voltage, if the device says
	Reading    *Environment
	Presses    []device.ButtonPress
}

// IsEmpty is whether the advertisement said nothing the plugin keeps.
func (r Result) IsEmpty() bool {
	return r.Battery == nil && r.Millivolts == nil && r.Reading == nil && len(r.Presses) == 0
}

// Percent is a pointer to a battery percentage, for Result.Battery.
func Percent(level int) *int {
	return &level
}

// Device is one device. It owns that device's state: the Inkbird's changing keys, a button's burst.
//
// Hear runs on the bluetooth plugin's goroutine, one advertisement at a time. It must be fast and must not block or keep the
// advertisement. Anything heard from a claimed device also counts as "seen", with its RSSI and source, whatever the Result says.
type Device interface {
	Hear(a *AdvertisementEvent, now time.Time) Result
}

// DeviceModel recognises a kind of device nobody configured, and decodes it.
type DeviceModel interface {
	Name() string               // "Govee GVH5075"
	Kind() Kind                 // what entity it is
	Rules() []Rule              // what sources should forward, and how fast
	NeedsScanResponse() bool    // its data is in scan responses
	Identify(known Fields) bool // from everything heard of a device so far
	New() Device
}

// ConfiguredDevice is a device the assembly configured. Usually it is claimed by a fixed address (see Addressed); a phone's
// changes, so claiming is a question.
type ConfiguredDevice interface {
	Name() string
	Kind() Kind
	Claims(a *AdvertisementEvent) bool // asked once per new address, then cached
	Rules() []Rule                     // sources forward what these match
	NeedsScanResponse() bool
	New() Device
}

// Addressed is implemented by a ConfiguredDevice that has one fixed address, in the form NormalizeAddress makes. Its entity
// is then registered at the start, before it is heard.
type Addressed interface {
	Address() string
}

// MakeAndModeler is implemented by a ConfiguredDevice that can say what it is: "Inkbird IBS-TH".
type MakeAndModeler interface {
	MakeAndModel() string
}

// Validator is implemented by a ConfiguredDevice that can say what is wrong with how it was configured, which the plugin
// refuses to start with.
type Validator interface {
	Validate() error
}

// ButtonBinder is implemented by a ConfiguredDevice that is also the handle to its button entity (see ButtonHandle). The plugin
// binds it when it registers the entity, and unbinds it when it deregisters it.
type ButtonBinder interface {
	Bind(container spi.IPMAASContainer, id string, event common.StubAvailableEvent[device.IButton]) error
	Unbind()
}

// ButtonHandle is the assembly's side of a button entity: embed it in the ConfiguredDevice a builder returns, and the assembly
// can use that as the button (device.IButton) from the start. Before the plugin has registered the entity, and after it stops,
// it answers as a button that has not been pressed.
type ButtonHandle struct {
	name     string
	stub     common.StubHolder[device.IButton]
	entityId atomic.Pointer[string]
}

var _ device.IButton = (*ButtonHandle)(nil)
var _ ButtonBinder = (*ButtonHandle)(nil)

func NewButtonHandle(name string) *ButtonHandle {
	return &ButtonHandle{name: name}
}

// EntityId is the ID the button is registered under once the plugin has started, which is what the events about it say it is
// from, or empty before then.
func (h *ButtonHandle) EntityId() string {
	if id := h.entityId.Load(); id != nil {
		return *id
	}

	return ""
}

func (h *ButtonHandle) GetButtonData() device.Button {
	stub, err := h.stub.Stub()
	if err != nil {
		return device.Button{Name: h.name}
	}

	return stub.GetButtonData()
}

// AddOnStubAvailableListener adds a function to be called when the button's stub becomes available (see common.StubHolder).
func (h *ButtonHandle) AddOnStubAvailableListener(listener func(event common.StubAvailableEvent[device.IButton])) {
	h.stub.AddOnStubAvailableListener(listener)
}

// Bind is for the plugin.
func (h *ButtonHandle) Bind(container spi.IPMAASContainer, id string, event common.StubAvailableEvent[device.IButton]) error {
	h.entityId.Store(&id)

	return h.stub.Provide(container, event)
}

// Unbind is for the plugin.
func (h *ButtonHandle) Unbind() {
	h.stub.Withdraw()
	h.entityId.Store(nil)
}

// ConfigureModel is the device at the address, which is of the model: a name for it, and its model without having to
// recognise it. It is what a builder in a device package returns for a device whose address is fixed. It panics if the address
// is not a Bluetooth address, like regexp.MustCompile does for a pattern that can't be: it is written in the program.
func ConfigureModel(model DeviceModel, address string, name string) ConfiguredDevice {
	normal, ok := NormalizeAddress(address)
	if !ok {
		panic(fmt.Sprintf("%s address %q is not a Bluetooth address like AA:BB:CC:DD:EE:FF", model.Name(), address))
	}

	return &configuredModel{model: model, address: normal, name: name}
}

type configuredModel struct {
	model   DeviceModel
	address string
	name    string
}

var _ ConfiguredDevice = (*configuredModel)(nil)
var _ Addressed = (*configuredModel)(nil)
var _ MakeAndModeler = (*configuredModel)(nil)

func (c *configuredModel) Name() string                      { return c.name }
func (c *configuredModel) Kind() Kind                        { return c.model.Kind() }
func (c *configuredModel) Address() string                   { return c.address }
func (c *configuredModel) MakeAndModel() string              { return c.model.Name() }
func (c *configuredModel) Claims(a *AdvertisementEvent) bool { return a.Address == c.address }
func (c *configuredModel) NeedsScanResponse() bool           { return c.model.NeedsScanResponse() }
func (c *configuredModel) New() Device                       { return c.model.New() }
func (c *configuredModel) Rules() []Rule {
	return []Rule{{Name: "configured " + c.address, Address: c.address}}
}
