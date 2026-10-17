package bluetooth

import (
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	"github.com/avanha/pmaas-spi/events"
)

func sightingsEvery(t *testing.T, interval time.Duration) (*Interests, *SightingTracker) {
	t.Helper()

	interests := NewInterests()
	if err := interests.Set(InterestEvent{Owner: "bluetooth", Sightings: interval}); err != nil {
		t.Fatal(err)
	}

	return interests, NewSightingTracker(interests)
}

func heardOf(address string, name string, rssi int) *AdvertisementEvent {
	return &AdvertisementEvent{Source: "src", Address: address, AddressType: AddressRandom, RSSI: rssi, Time: time.Now(), Fields: Fields{LocalName: name}}
}

func TestSightingTracker_NothingIsToldUntilSomebodyAsksForSightings(t *testing.T) {
	tracker := NewSightingTracker(NewInterests())

	if _, told := tracker.Observe(heardOf("44:DC:FA:B6:5B:00", "Phone", -60)); told {
		t.Error("told with nobody asking")
	}

	// Asking for advertisements is not asking for sightings.
	interests := NewInterests()
	_ = interests.Set(InterestEvent{Owner: "bluetooth", Rules: []Rule{{NamePrefix: "x"}}})

	if _, told := NewSightingTracker(interests).Observe(heardOf("44:DC:FA:B6:5B:00", "Phone", -60)); told {
		t.Error("told of sightings nobody asked for")
	}
}

func TestSightingTracker_ADeviceIsToldOfAtOnceAndThenOncePerInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		_, tracker := sightingsEvery(t, time.Minute)

		first, told := tracker.Observe(heardOf("44:DC:FA:B6:5B:00", "", -60))
		if !told || first.Address != "44:DC:FA:B6:5B:00" || first.Source != "src" || first.RSSI != -60 || first.AddressType != AddressRandom {
			t.Fatalf("%v %+v", told, first)
		}

		for range 5 {
			time.Sleep(10 * time.Second)

			if _, told := tracker.Observe(heardOf("44:DC:FA:B6:5B:00", "", -70)); told {
				t.Fatal("told again within the interval")
			}
		}

		time.Sleep(10 * time.Second)

		again, told := tracker.Observe(heardOf("44:DC:FA:B6:5B:00", "", -75))
		if !told || again.RSSI != -75 {
			t.Errorf("%v %+v", told, again)
		}

		// Another device is its own.
		if _, told := tracker.Observe(heardOf("44:DC:FA:B6:5B:01", "", -60)); !told {
			t.Error("a different device was not told of")
		}
	})
}

func TestSightingTracker_ANameLearnedLaterIsToldAtOnceAndRememberedAfter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		interests, tracker := sightingsEvery(t, time.Hour)

		if first, _ := tracker.Observe(heardOf("A4:C1:38:00:00:01", "", -60)); first.LocalName != "" {
			t.Fatalf("a name from nowhere: %+v", first)
		}

		named, told := tracker.Observe(heardOf("A4:C1:38:00:00:01", "Basement", -60))
		if !told || named.LocalName != "Basement" {
			t.Fatalf("a name learned was not told at once: %v %+v", told, named)
		}

		// The same name again is not news. After the interval the name is still told, though an advertisement carried none.
		if _, told := tracker.Observe(heardOf("A4:C1:38:00:00:01", "Basement", -60)); told {
			t.Error("told of a name already told")
		}

		time.Sleep(time.Hour)
		_ = interests.Set(InterestEvent{Owner: "bluetooth", Sightings: time.Hour}) // said again, as it is every minute

		later, told := tracker.Observe(heardOf("A4:C1:38:00:00:01", "", -60))
		if !told || later.LocalName != "Basement" {
			t.Errorf("%v %+v", told, later)
		}
	})
}

func TestSightingTracker_ServicesAndCompaniesAreWhatHasBeenSeenOverTime(t *testing.T) {
	_, tracker := sightingsEvery(t, time.Hour)

	first := heardOf("A4:C1:38:00:00:01", "x", -60)
	first.ServiceUUIDs = []string{UUID16(0xec88)}
	first.ManufacturerData = map[uint16][]byte{0x004c: {2}}
	tracker.Observe(first)

	second := heardOf("A4:C1:38:00:00:01", "y", -60)
	second.ServiceUUIDs = []string{UUID16(0xec88), UUID16(0x180f)}
	second.ManufacturerData = map[uint16][]byte{0xec88: {1}, 0x0001: nil}

	// Not told on the interval, but remembered: the next that is told has it all.
	tracker.Observe(second)

	third := heardOf("A4:C1:38:00:00:01", "", -60)
	third.Time = time.Now()

	interests := NewInterests()
	_ = interests.Set(InterestEvent{Owner: "o", Sightings: time.Nanosecond})
	tracker.interests = interests

	time.Sleep(time.Millisecond)

	got, told := tracker.Observe(third)
	if !told {
		t.Fatal("not told")
	}

	if !reflect.DeepEqual(got.ServiceUUIDs, []string{UUID16(0xec88), UUID16(0x180f)}) ||
		!reflect.DeepEqual(got.ManufacturerIDs, []uint16{0x0001, 0x004c, 0xec88}) {
		t.Errorf("%+v", got)
	}
}

func TestSightingTracker_TheShortestIntervalAnybodyAsksForGoverns(t *testing.T) {
	interests := NewInterests()
	_ = interests.Set(InterestEvent{Owner: "a", Sightings: time.Hour})
	_ = interests.Set(InterestEvent{Owner: "b", Sightings: time.Minute})
	_ = interests.Set(InterestEvent{Owner: "c", Rules: []Rule{{NamePrefix: "x"}}})

	// The owners are kept in a map, whose order is not the same twice.
	for range 50 {
		if got := interests.SightingInterval(); got != time.Minute {
			t.Fatalf("interval %v", got)
		}
	}

	_ = interests.Set(InterestEvent{Owner: "b"})

	if got := interests.SightingInterval(); got != time.Hour {
		t.Errorf("interval %v after b left", got)
	}

	_ = interests.Set(InterestEvent{Owner: "a"})

	if got := interests.SightingInterval(); got != 0 {
		t.Errorf("interval %v with nobody asking", got)
	}
}

func TestSightingTracker_WhatItRemembersIsBounded(t *testing.T) {
	_, tracker := sightingsEvery(t, time.Hour)

	for i := 0; i < maxSeen+10; i++ {
		tracker.Observe(heardOf(FormatAddress(uint64(0x440000000000+i)), "", -60))
	}

	if len(tracker.seen) > maxSeen {
		t.Errorf("remembers %d devices", len(tracker.seen))
	}
}

func TestSightings_AreInterestOnTheirOwnAndAreValidated(t *testing.T) {
	interests := NewInterests()

	if err := interests.Set(InterestEvent{Owner: "a", Sightings: time.Minute}); err != nil {
		t.Fatal(err)
	}

	if !interests.Interested() || len(interests.Rules()) != 0 {
		t.Error("sightings alone are an interest, with no rules")
	}

	if err := interests.Set(InterestEvent{Owner: "a", Sightings: -time.Second}); err == nil {
		t.Error("a negative interval was accepted")
	}

	// What was there stands.
	if interests.SightingInterval() != time.Minute {
		t.Error("what was there did not stand")
	}
}

func TestPublishSightingAndInterestEvent(t *testing.T) {
	b := &recording{}

	if err := PublishSighting(b, SightingEvent{Source: "src", Address: "A4:C1:38:00:00:01"}); err != nil || b.ids[0] != "src" {
		t.Errorf("%v %v", err, b.ids)
	}

	if err := PublishInterestEvent(b, InterestEvent{Owner: "bluetooth", Sightings: time.Minute}); err != nil || b.ids[1] != "bluetooth" {
		t.Errorf("%v %v", err, b.ids)
	}

	if err := PublishInterestEvent(b, InterestEvent{Sightings: time.Minute}); err == nil || len(b.events) != 2 {
		t.Error("an interest with no owner was sent")
	}

	if !SightingPredicate(&events.EventInfo{Event: SightingEvent{}}) || SightingPredicate(&events.EventInfo{Event: "x"}) ||
		AdvertisementPredicate(&events.EventInfo{Event: SightingEvent{}}) {
		t.Error("predicate")
	}
}
