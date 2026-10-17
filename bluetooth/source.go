package bluetooth

import (
	"reflect"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

// ScanMode is what a source's radio is doing.
type ScanMode int

const (
	ScanOff ScanMode = iota
	ScanPassive
	ScanActive
)

func (m ScanMode) String() string {
	switch m {
	case ScanPassive:
		return "passive"
	case ScanActive:
		return "active"
	default:
		return "off"
	}
}

// Source is a place advertisements are heard, as its entity reports it.
type Source struct {
	Name string

	// Kind is "bluez" for the local radio, or "esphome-proxy".
	Kind string

	// Connectivity is the transport's: the proxy's connection, or the adapter being up. For a proxy it is the ESPHome
	// device's, so an offline alert on the device covers it.
	Connectivity environment.Connectivity
	OnlineSince  time.Time
	OfflineSince time.Time

	Scanning ScanMode

	// LastHeard is when the source last heard an advertisement at all, wanted or not.
	LastHeard time.Time

	// Heard is every advertisement, Forwarded those broadcast, and Dropped those that could not be decoded or whose
	// broadcast failed. The difference between Heard and Forwarded is what the interest rules and coalescing saved.
	Heard     uint64
	Forwarded uint64
	Dropped   uint64
}

// ISource is the entity a source registers. It only reports in v1. When something has to write to Bluetooth, the methods
// go here (connect, read, write), and not on the event bus.
type ISource interface {
	GetSourceData() Source
}

// ISourceType is the entity type to register a source under.
var ISourceType = reflect.TypeFor[ISource]()
