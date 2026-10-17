package bluetooth

import (
	"reflect"
	"sync"
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

	// Sighted is the devices told of in sightings, which are not counted in Forwarded.
	Sighted uint64
}

// ISource is the entity a source registers. It only reports in v1. When something has to write to Bluetooth, the methods
// go here (connect, read, write), and not on the event bus.
type ISource interface {
	GetSourceData() Source
}

// ISourceType is the entity type to register a source under.
var ISourceType = reflect.TypeFor[ISource]()

// SourceState is the entity a source registers, kept up to date by whatever hears the advertisements and read by anything
// else. It is safe to use from more than one goroutine, and the stub a source's entity registration hands out can be the
// SourceState itself.
type SourceState struct {
	mu   sync.Mutex
	data Source
}

var _ ISource = (*SourceState)(nil)

// NewSourceState makes the state of a source that is not yet known to be connected to anything.
func NewSourceState(name string, kind string) *SourceState {
	return &SourceState{data: Source{Name: name, Kind: kind}}
}

func (s *SourceState) GetSourceData() Source {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data
}

// SetConnectivity notes whether the source's transport is up. The time it has been so is only moved when that changes.
func (s *SourceState) SetConnectivity(connectivity environment.Connectivity, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data.Connectivity == connectivity {
		return
	}

	s.data.Connectivity = connectivity

	switch connectivity {
	case environment.ConnectivityOnline:
		s.data.OnlineSince = at
	case environment.ConnectivityOffline:
		s.data.OfflineSince = at
	}
}

// SetScanning notes what the radio is doing.
func (s *SourceState) SetScanning(mode ScanMode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Scanning = mode
}

// Heard notes an advertisement, wanted or not.
func (s *SourceState) Heard(at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Heard++
	s.data.LastHeard = at
}

// Forwarded notes an advertisement that was broadcast.
func (s *SourceState) Forwarded() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Forwarded++
}

// Sighted notes a sighting that was broadcast.
func (s *SourceState) Sighted() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Sighted++
}

// Dropped notes an advertisement that could not be decoded or broadcast.
func (s *SourceState) Dropped() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Dropped++
}
