package bluetooth

import (
	"sync"
	"time"
)

// Verdict is what a Coalescer decided about an advertisement.
type Verdict int

const (
	// NotWanted is that no rule matches, or nobody has said what they want yet.
	NotWanted Verdict = iota

	// Duplicate is that a rule matches but the same address said the same thing a moment ago.
	Duplicate

	// Forward is that the advertisement should be broadcast.
	Forward
)

func (v Verdict) String() string {
	switch v {
	case Duplicate:
		return "duplicate"
	case Forward:
		return "forward"
	default:
		return "not wanted"
	}
}

// maxTracked is how many addresses a Coalescer remembers the last thing from. Past it, the old ones are dropped, which at
// worst forwards an identical repeat once more.
const maxTracked = 4096

// pruneAfter is how long an address is remembered for when the memory is full.
const pruneAfter = 5 * time.Minute

// Interests is what consumers have said they want, with each owner's interest expiring after InterestTTL unless it is said
// again. A source with more than one place it hears from (a Bluetooth proxy for each of several devices) keeps one, and
// gives it to each Coalescer.
//
// It is safe to use from more than one goroutine.
type Interests struct {
	mu     sync.Mutex
	now    func() time.Time
	owners map[string]ownerInterest
}

type ownerInterest struct {
	rules      []Rule
	activeScan bool
	heard      time.Time
}

// NewInterests makes an Interests that nobody has said anything to.
func NewInterests() *Interests {
	return &Interests{now: time.Now, owners: make(map[string]ownerInterest)}
}

// Set takes in an owner's whole current interest, replacing what it said before. An interest that is not valid is refused
// and what the owner said before stands.
func (i *Interests) Set(event InterestEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if len(event.Rules) == 0 {
		delete(i.owners, event.Owner)

		return nil
	}

	i.owners[event.Owner] = ownerInterest{
		rules:      append([]Rule(nil), event.Rules...),
		activeScan: event.ActiveScan,
		heard:      i.now(),
	}

	return nil
}

// Interested is whether anybody currently wants anything.
func (i *Interests) Interested() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	return len(i.live()) > 0
}

// ActiveScan is whether anybody currently wants scan responses.
func (i *Interests) ActiveScan() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, owner := range i.live() {
		if owner.activeScan {
			return true
		}
	}

	return false
}

// Rules is every rule anybody currently wants, for a source that can pass some of them on to the radio.
func (i *Interests) Rules() []Rule {
	i.mu.Lock()
	defer i.mu.Unlock()

	var rules []Rule

	for _, owner := range i.live() {
		rules = append(rules, owner.rules...)
	}

	return rules
}

// snapshot is the owners whose interest has not expired.
func (i *Interests) snapshot() []ownerInterest {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.live()
}

// live is the owners whose interest has not expired, forgetting those that have. Called with the lock held.
func (i *Interests) live() []ownerInterest {
	now := i.now()
	live := make([]ownerInterest, 0, len(i.owners))

	for name, owner := range i.owners {
		if now.Sub(owner.heard) >= InterestTTL {
			delete(i.owners, name)

			continue
		}

		live = append(live, owner)
	}

	return live
}

// Coalescer is what every source uses to apply what consumers want, so that they all treat an advertisement the same way:
//
//   - Nothing matching no rule is forwarded, and nothing at all until somebody has said what they want.
//   - Immediate: forwarded, always.
//   - Coalesce: the first with a given address and payload is forwarded at once (the leading edge), and identical repeats
//     within the rule's window are not. After the window an identical one is forwarded again, so that a device which never
//     changes still shows as alive. The signal strength is not part of "identical".
//   - If more than one rule matches, Immediate wins, and otherwise the shortest window.
//
// It is safe to use from more than one goroutine: a source feeds interests from one and advertisements from another. A source
// that hears from several places gives each its own Coalescer over one Interests, so each place's repeats are dropped
// separately and every place is heard.
type Coalescer struct {
	mu        sync.Mutex
	now       func() time.Time
	interests *Interests
	last      map[string]forwarded
}

type forwarded struct {
	key string
	at  time.Time
}

// NewCoalescer makes a Coalescer that has not been told what anybody wants.
func NewCoalescer() *Coalescer {
	return NewCoalescerFor(NewInterests())
}

// NewCoalescerFor makes a Coalescer that applies the interests. Others can share them.
func NewCoalescerFor(interests *Interests) *Coalescer {
	return &Coalescer{now: interests.now, interests: interests, last: make(map[string]forwarded)}
}

// SetInterest takes in an owner's whole current interest. See Interests.Set.
func (c *Coalescer) SetInterest(event InterestEvent) error {
	return c.interests.Set(event)
}

// Interested is whether anybody currently wants anything.
func (c *Coalescer) Interested() bool {
	return c.interests.Interested()
}

// ActiveScan is whether anybody currently wants scan responses.
func (c *Coalescer) ActiveScan() bool {
	return c.interests.ActiveScan()
}

// Rules is every rule anybody currently wants.
func (c *Coalescer) Rules() []Rule {
	return c.interests.Rules()
}

// Admit decides what to do about an advertisement. The rule that matched is returned for a source to count by, nil if none.
func (c *Coalescer) Admit(a *AdvertisementEvent) (Verdict, *Rule) {
	return c.AdmitAs(a, a)
}

// AdmitAs is Admit for a source that knows more about the device than one advertisement says, as BlueZ does: the rules are
// matched against everything known about the device (view), while what is forwarded, and what counts as a repeat, is the
// advertisement.
func (c *Coalescer) AdmitAs(view *AdvertisementEvent, a *AdvertisementEvent) (Verdict, *Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()

	rule := c.choose(view)
	if rule == nil {
		return NotWanted, nil
	}

	now := c.now()

	if rule.Latency == Immediate {
		c.remember(a, now)

		return Forward, rule
	}

	if previous, seen := c.last[a.Address]; seen && previous.key == a.PayloadKey() && now.Sub(previous.at) < rule.window() {
		return Duplicate, rule
	}

	c.remember(a, now)

	return Forward, rule
}

// choose is the rule that governs the advertisement, or nil. Called with the lock held.
func (c *Coalescer) choose(a *AdvertisementEvent) *Rule {
	var chosen *Rule

	for _, owner := range c.interests.snapshot() {
		for i := range owner.rules {
			rule := &owner.rules[i]
			if !rule.Matches(a) {
				continue
			}

			if chosen == nil || rule.Latency == Immediate && chosen.Latency != Immediate ||
				rule.Latency == Coalesce && chosen.Latency == Coalesce && rule.window() < chosen.window() {
				chosen = rule
			}
		}
	}

	return chosen
}

// remember notes what was last forwarded for the address. Called with the lock held.
func (c *Coalescer) remember(a *AdvertisementEvent, now time.Time) {
	if len(c.last) >= maxTracked {
		for address, previous := range c.last {
			if now.Sub(previous.at) >= pruneAfter {
				delete(c.last, address)
			}
		}

		// Still full of recent ones: start again, which can only cost a repeat being forwarded.
		if len(c.last) >= maxTracked {
			c.last = make(map[string]forwarded)
		}
	}

	c.last[a.Address] = forwarded{key: a.PayloadKey(), at: now}
}
