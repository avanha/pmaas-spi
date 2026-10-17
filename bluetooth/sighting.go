package bluetooth

import (
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/avanha/pmaas-spi/events"
)

// SightingEvent is broadcast by a source to say that it has heard a device that nobody has asked for the advertisements of,
// when somebody has asked to be told of the devices around (InterestEvent.Sightings). It has enough to say what the device is
// called and what kind it is, and none of what it said, so the many devices that are only passing by cost little. It is how a
// list of the devices that are around can be kept, to find the address of one that should be configured.
type SightingEvent struct {
	// Source is the entity ID of the source that heard it.
	Source string

	// Address is "A4:C1:38:CC:E2:F0": upper case, colon separated.
	Address     string
	AddressType AddressType

	// RSSI is the signal strength in dBm, as this source last heard it.
	RSSI int

	// Time is when the source last heard it.
	Time time.Time

	// LocalName is the name the device has advertised, if the source has seen one. A name often comes in a scan response, so
	// a source remembers it from one advertisement to the next.
	LocalName string

	// ServiceUUIDs are the services the device has listed (128-bit, lower case), and ManufacturerIDs the companies whose
	// manufacturer data it carries, which says who made it. Both are what the source has seen of the device, not only in the
	// last advertisement.
	ServiceUUIDs    []string
	ManufacturerIDs []uint16
}

// SightingPredicate is an events.EventPredicate matching SightingEvents, for a consumer.
func SightingPredicate(eventInfo *events.EventInfo) bool {
	_, ok := eventInfo.Event.(SightingEvent)

	return ok
}

// PublishSighting broadcasts the sighting from its source.
func PublishSighting(broadcaster Broadcaster, sighting SightingEvent) error {
	return broadcaster.BroadcastEvent(sighting.Source, sighting)
}

// maxCompanies is how many manufacturer IDs are kept for a device. Some devices put a reading where the company ID goes, so
// the number of different ones they show grows without end.
const maxCompanies = 16

// maxSeen is how many devices a SightingTracker remembers. Past it, the ones not heard from for longest are let go.
const maxSeen = 4096

// SightingTracker is what a source uses to decide when to tell of a device that is not otherwise wanted: when it is first
// heard, when its name is first learned (which is what a person looking for it wants to see), and then every
// InterestEvent.Sightings. It remembers what a device has said over time, because it is the sum of that which identifies it.
//
// A source with several places it hears from makes one for each, over the same Interests. It is safe to use from more than
// one goroutine.
type SightingTracker struct {
	mu        sync.Mutex
	now       func() time.Time
	interests *Interests
	seen      map[string]*seenDevice
}

type seenDevice struct {
	name         string
	uuids        []string
	manufacturer map[uint16]bool
	reported     time.Time
	heard        time.Time
}

// NewSightingTracker makes a tracker that tells of devices as often as the interests ask.
func NewSightingTracker(interests *Interests) *SightingTracker {
	return &SightingTracker{now: interests.now, interests: interests, seen: make(map[string]*seenDevice)}
}

// Observe takes in an advertisement, or a source's whole knowledge of a device, and says whether to tell of it now, with the
// sighting to broadcast. Nothing is told when nobody has asked for sightings.
func (t *SightingTracker) Observe(a *AdvertisementEvent) (SightingEvent, bool) {
	interval := t.interests.SightingInterval()
	if interval <= 0 {
		return SightingEvent{}, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()

	d, known := t.seen[a.Address]
	if !known {
		if len(t.seen) >= maxSeen {
			t.dropOldest()
		}

		d = &seenDevice{manufacturer: make(map[uint16]bool)}
		t.seen[a.Address] = d
	}

	d.heard = now

	// A name learned for the first time is told at once; so is a device never told of.
	learnedName := a.LocalName != "" && d.name == ""
	if a.LocalName != "" {
		d.name = a.LocalName
	}

	for _, uuid := range a.ServiceUUIDs {
		if !slices.Contains(d.uuids, uuid) {
			d.uuids = append(d.uuids, uuid)
		}
	}

	for id := range a.ManufacturerData {
		if len(d.manufacturer) < maxCompanies {
			d.manufacturer[id] = true
		}
	}

	if known && !learnedName && now.Sub(d.reported) < interval {
		return SightingEvent{}, false
	}

	d.reported = now

	ids := make([]int, 0, len(d.manufacturer))
	for id := range d.manufacturer {
		ids = append(ids, int(id))
	}

	sort.Ints(ids)

	manufacturers := make([]uint16, len(ids))
	for i, id := range ids {
		manufacturers[i] = uint16(id)
	}

	return SightingEvent{
		Source:          a.Source,
		Address:         a.Address,
		AddressType:     a.AddressType,
		RSSI:            a.RSSI,
		Time:            a.Time,
		LocalName:       d.name,
		ServiceUUIDs:    slices.Clone(d.uuids),
		ManufacturerIDs: manufacturers,
	}, true
}

// dropOldest lets go of the device heard from least recently. Called with the lock held.
func (t *SightingTracker) dropOldest() {
	var oldest string
	var oldestHeard time.Time

	for address, d := range t.seen {
		if oldest == "" || d.heard.Before(oldestHeard) {
			oldest, oldestHeard = address, d.heard
		}
	}

	delete(t.seen, oldest)
}
