package bluetooth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/avanha/pmaas-spi/events"
)

func govee() *AdvertisementEvent {
	return &AdvertisementEvent{
		Address: "A4:C1:38:CC:E2:F0",
		Fields: Fields{
			LocalName:        "GVH5075_E2F0",
			ServiceUUIDs:     []string{UUID16(0xec88)},
			ManufacturerData: map[uint16][]byte{0xec88: {0x00, 0x03, 0x39}, 0x004c: {0x02, 0x15}},
			ServiceData:      map[string][]byte{UUID16(0xfcd2): {1}},
		},
	}
}

func TestRule_Matches(t *testing.T) {
	a := govee()

	for name, c := range map[string]struct {
		rule Rule
		want bool
	}{
		"address":                       {Rule{Address: "A4:C1:38:CC:E2:F0"}, true},
		"another address":               {Rule{Address: "A4:C1:38:CC:E2:F1"}, false},
		"name prefix":                   {Rule{NamePrefix: "GVH"}, true},
		"another name":                  {Rule{NamePrefix: "sps"}, false},
		"manufacturer":                  {Rule{ManufacturerID: ManufacturerOf(0xec88)}, true},
		"another manufacturer":          {Rule{ManufacturerID: ManufacturerOf(0x0766)}, false},
		"manufacturer data prefix":      {Rule{ManufacturerID: ManufacturerOf(0xec88), ManufacturerPrefix: []byte{0x00, 0x03}}, true},
		"another manufacturer data":     {Rule{ManufacturerID: ManufacturerOf(0xec88), ManufacturerPrefix: []byte{0x01}}, false},
		"prefix longer than the data":   {Rule{ManufacturerID: ManufacturerOf(0xec88), ManufacturerPrefix: []byte{0, 3, 0x39, 0}}, false},
		"listed uuid":                   {Rule{ServiceUUID: UUID16(0xec88)}, true},
		"uuid of service data":          {Rule{ServiceUUID: UUID16(0xfcd2)}, true},
		"another uuid":                  {Rule{ServiceUUID: UUID16(0xfff0)}, false},
		"all of it":                     {Rule{Address: "A4:C1:38:CC:E2:F0", NamePrefix: "GVH", ManufacturerID: ManufacturerOf(0xec88)}, true},
		"all but one":                   {Rule{Address: "A4:C1:38:CC:E2:F0", NamePrefix: "sps", ManufacturerID: ManufacturerOf(0xec88)}, false},
		"a manufacturer with no data":   {Rule{ManufacturerID: ManufacturerOf(0x004c)}, true},
		"an advertisement with no name": {Rule{NamePrefix: "GVH"}, true},
	} {
		if got := c.rule.Matches(a); got != c.want {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}

	if (Rule{NamePrefix: "GVH"}).Matches(&AdvertisementEvent{Address: "A4:C1:38:CC:E2:F0"}) {
		t.Error("a name prefix matched an advertisement with no name")
	}
}

func TestRule_Validate(t *testing.T) {
	good := []Rule{
		{Address: "A4:C1:38:CC:E2:F0"},
		{NamePrefix: "sps", ServiceUUID: UUID16(0xfff0)},
		{ManufacturerID: ManufacturerOf(0), ManufacturerPrefix: []byte{1}, Latency: Immediate},
		{NamePrefix: "x", Window: time.Second},
	}

	for _, r := range good {
		if err := r.Validate(); err != nil {
			t.Errorf("%+v: %v", r, err)
		}
	}

	bad := map[string]Rule{
		"empty":                  Rule{Name: "everything"},
		"address in lower":       {Address: "a4:c1:38:cc:e2:f0"},
		"address not an address": {Address: "nope"},
		"prefix without id":      {NamePrefix: "x", ManufacturerPrefix: []byte{1}},
		"short uuid":             {ServiceUUID: "ec88"},
		"upper case uuid":        {ServiceUUID: strings.ToUpper(UUID16(0xec88))},
		"negative window":        {NamePrefix: "x", Window: -time.Second},
	}

	for name, r := range bad {
		if r.Validate() == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

func TestInterestEvent_Validate(t *testing.T) {
	if err := (InterestEvent{Owner: "bluetooth", Rules: []Rule{{NamePrefix: "x"}}}).Validate(); err != nil {
		t.Error(err)
	}

	if (InterestEvent{Rules: []Rule{{NamePrefix: "x"}}}).Validate() == nil {
		t.Error("an interest with no owner")
	}

	err := (InterestEvent{Owner: "o", Rules: []Rule{{NamePrefix: "x"}, {Name: "bad"}}}).Validate()
	if err == nil || !strings.Contains(err.Error(), "rule 1") || !strings.Contains(err.Error(), `"bad"`) {
		t.Errorf("error %v", err)
	}

	// Nothing wanted is fine: it withdraws the owner.
	if err := (InterestEvent{Owner: "o"}).Validate(); err != nil {
		t.Error(err)
	}
}

type recording struct {
	ids    []string
	events []any
	err    error
}

func (r *recording) BroadcastEvent(id string, event any) error {
	r.ids = append(r.ids, id)
	r.events = append(r.events, event)

	return r.err
}

func TestPublishInterest_AndAdvertisement(t *testing.T) {
	b := &recording{}

	if err := PublishInterest(b, "bluetooth", []Rule{{NamePrefix: "x"}}, true); err != nil {
		t.Fatal(err)
	}

	event, ok := b.events[0].(InterestEvent)
	if !ok || b.ids[0] != "bluetooth" || event.Owner != "bluetooth" || !event.ActiveScan || len(event.Rules) != 1 {
		t.Errorf("%v %+v", b.ids, b.events)
	}

	// An interest that is not valid is not sent.
	if err := PublishInterest(b, "bluetooth", []Rule{{}}, false); err == nil || len(b.events) != 1 {
		t.Errorf("sent a bad interest: %v, %d events", err, len(b.events))
	}

	if err := PublishAdvertisement(b, AdvertisementEvent{Source: "esphome_1", Address: "A4:C1:38:CC:E2:F0"}); err != nil {
		t.Fatal(err)
	}

	if b.ids[1] != "esphome_1" {
		t.Errorf("sent from %q", b.ids[1])
	}

	b.err = errors.New("nope")
	if PublishAdvertisement(b, AdvertisementEvent{}) == nil {
		t.Error("a failure to broadcast was lost")
	}
}

func TestPredicates(t *testing.T) {
	for _, c := range []struct {
		event                   any
		advertisement, interest bool
	}{
		{AdvertisementEvent{}, true, false},
		{InterestEvent{}, false, true},
		{"something else", false, false},
		{nil, false, false},
	} {
		info := &events.EventInfo{Event: c.event}
		if AdvertisementPredicate(info) != c.advertisement || InterestPredicate(info) != c.interest {
			t.Errorf("%T: advertisement %v, interest %v", c.event, AdvertisementPredicate(info), InterestPredicate(info))
		}
	}
}

func TestPayloadKey(t *testing.T) {
	a, b := govee(), govee()
	b.RSSI, b.Source, b.Time = -40, "elsewhere", time.Now()

	if a.PayloadKey() != b.PayloadKey() {
		t.Error("signal strength, source or time changed the key")
	}

	for name, change := range map[string]func(*AdvertisementEvent){
		"name":              func(x *AdvertisementEvent) { x.LocalName = "other" },
		"uuids":             func(x *AdvertisementEvent) { x.ServiceUUIDs = append(x.ServiceUUIDs, UUID16(1)) },
		"manufacturer data": func(x *AdvertisementEvent) { x.ManufacturerData[0xec88] = []byte{9} },
		"manufacturer id":   func(x *AdvertisementEvent) { x.ManufacturerData[0x0001] = nil },
		"service data":      func(x *AdvertisementEvent) { x.ServiceData[UUID16(0xfcd2)] = []byte{2} },
		"tx power":          func(x *AdvertisementEvent) { p := int8(-4); x.TxPower = &p },
	} {
		c := govee()
		change(c)

		if c.PayloadKey() == a.PayloadKey() {
			t.Errorf("%s did not change the key", name)
		}
	}

	// The order of the lists and maps does not matter.
	x := &AdvertisementEvent{Fields: Fields{ServiceUUIDs: []string{"b", "a"}}}
	y := &AdvertisementEvent{Fields: Fields{ServiceUUIDs: []string{"a", "b"}}}

	if x.PayloadKey() != y.PayloadKey() {
		t.Error("the order of UUIDs changed the key")
	}

	// Raw data is what it says, when there is any.
	r1 := &AdvertisementEvent{Raw: []byte{1, 2}, Fields: Fields{LocalName: "a"}}
	r2 := &AdvertisementEvent{Raw: []byte{1, 2}, Fields: Fields{LocalName: "b"}}
	r3 := &AdvertisementEvent{Raw: []byte{1, 3}}

	if r1.PayloadKey() != r2.PayloadKey() || r1.PayloadKey() == r3.PayloadKey() {
		t.Error("raw data was not the key")
	}
}
