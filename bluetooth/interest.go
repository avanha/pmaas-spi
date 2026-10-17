package bluetooth

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/avanha/pmaas-spi/events"
)

// Latency is how eagerly a source forwards the advertisements a rule matches.
type Latency int

const (
	// Coalesce forwards the first advertisement with a given address and payload at once, and drops identical repeats
	// until the rule's window has passed. Right for sensors, which repeat themselves.
	Coalesce Latency = iota

	// Immediate forwards every advertisement the rule matches, as it comes, with nothing delayed, batched or dropped.
	// Right for a button, where each advertisement may be the press. The rule bounds the volume.
	Immediate
)

func (l Latency) String() string {
	if l == Immediate {
		return "immediate"
	}

	return "coalesce"
}

// DefaultWindow is how long identical repeats are dropped for, by a Coalesce rule that does not say.
const DefaultWindow = 30 * time.Second

// Rule says which advertisements are wanted, and how eagerly. Every field that is set has to match. An empty rule would
// match everything, which is never what is meant, so Validate refuses it.
type Rule struct {
	// Name is for logs and counters: "govee-h5075", "front-door-button".
	Name string

	// Address matches the advertisement's address exactly, in the form NormalizeAddress makes.
	Address string

	// NamePrefix matches the local name the advertisement carried.
	NamePrefix string

	// ManufacturerID matches when the advertisement carried manufacturer data for the company. ManufacturerPrefix, if
	// not empty, also requires that data to begin with those bytes.
	ManufacturerID     *uint16
	ManufacturerPrefix []byte

	// ServiceUUID matches when the advertisement listed the UUID, or carried service data for it.
	ServiceUUID string

	Latency Latency

	// Window is how long a Coalesce rule drops identical repeats for. Zero means DefaultWindow.
	Window time.Duration
}

// ManufacturerOf is a pointer to the company ID, for Rule.ManufacturerID.
func ManufacturerOf(id uint16) *uint16 {
	return &id
}

// Validate says what is wrong with the rule, if anything.
func (r Rule) Validate() error {
	if r.Address == "" && r.NamePrefix == "" && r.ManufacturerID == nil && r.ServiceUUID == "" {
		return errors.New("a rule has to match something: an empty one would match everything")
	}

	if r.Address != "" {
		if normal, ok := NormalizeAddress(r.Address); !ok || normal != r.Address {
			return fmt.Errorf("address %q is not in the form AA:BB:CC:DD:EE:FF (upper case)", r.Address)
		}
	}

	if len(r.ManufacturerPrefix) > 0 && r.ManufacturerID == nil {
		return errors.New("a manufacturer prefix needs a manufacturer ID")
	}

	if r.ServiceUUID != "" {
		if normal, ok := NormalizeUUID(r.ServiceUUID); !ok || normal != r.ServiceUUID {
			return fmt.Errorf("service UUID %q is not a 128-bit lower case UUID (see NormalizeUUID)", r.ServiceUUID)
		}
	}

	if r.Window < 0 {
		return errors.New("the window can't be negative")
	}

	return nil
}

// window is how long a Coalesce rule drops repeats for.
func (r Rule) window() time.Duration {
	if r.Window == 0 {
		return DefaultWindow
	}

	return r.Window
}

// Matches is whether the advertisement is wanted by the rule. A source that keeps whole devices, as BlueZ does, should
// match against the device as a whole, not against the part of it that an event carries.
func (r Rule) Matches(a *AdvertisementEvent) bool {
	if r.Address != "" && r.Address != a.Address {
		return false
	}

	if r.NamePrefix != "" && !strings.HasPrefix(a.LocalName, r.NamePrefix) {
		return false
	}

	if r.ManufacturerID != nil {
		data, ok := a.ManufacturerData[*r.ManufacturerID]
		if !ok || !bytes.HasPrefix(data, r.ManufacturerPrefix) {
			return false
		}
	}

	if r.ServiceUUID != "" {
		_, hasData := a.ServiceData[r.ServiceUUID]
		if !hasData && !slices.Contains(a.ServiceUUIDs, r.ServiceUUID) {
			return false
		}
	}

	return true
}

// InterestEvent is broadcast by a consumer to say what it wants to hear. It is the whole current set: it replaces
// everything its Owner sent before, and an empty set withdraws the owner's interest.
//
// A consumer broadcasts it when it starts, when it changes, and on every minute event, so a source that starts later, or
// restarts, catches up within a minute without anybody having to remember to ask. A source uses the union of what all
// owners want, and forgets an owner it has not heard from for InterestTTL, so a consumer that is removed stops costing
// anything. Until a source has heard from anybody, it forwards nothing.
type InterestEvent struct {
	// Owner is the consumer's name, so that two consumers do not replace each other.
	Owner string

	Rules []Rule

	// ActiveScan is whether any rule needs scan responses, which carry names for some devices. A source should scan
	// actively while any owner wants it.
	ActiveScan bool
}

// InterestTTL is how long an owner's interest lasts without being refreshed.
const InterestTTL = 3 * time.Minute

// Validate says what is wrong with the event, if anything.
func (e InterestEvent) Validate() error {
	if e.Owner == "" {
		return errors.New("an interest needs an owner")
	}

	for i, rule := range e.Rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("rule %d (%q): %w", i, rule.Name, err)
		}
	}

	return nil
}

// PublishInterest broadcasts the owner's whole current interest. The owner stands in for the entity ID.
func PublishInterest(broadcaster Broadcaster, owner string, rules []Rule, activeScan bool) error {
	event := InterestEvent{Owner: owner, Rules: rules, ActiveScan: activeScan}
	if err := event.Validate(); err != nil {
		return err
	}

	return broadcaster.BroadcastEvent(owner, event)
}

// InterestPredicate is an events.EventPredicate matching InterestEvents, for a source.
func InterestPredicate(eventInfo *events.EventInfo) bool {
	_, ok := eventInfo.Event.(InterestEvent)

	return ok
}
