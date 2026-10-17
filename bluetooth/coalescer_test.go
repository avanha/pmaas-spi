package bluetooth

import (
	"fmt"
	"testing"
	"testing/synctest"
	"time"
)

func button() Rule {
	return Rule{Name: "button", Address: "C3:00:00:00:00:01", Latency: Immediate}
}

func sensors() Rule {
	return Rule{Name: "govee", ManufacturerID: ManufacturerOf(0xec88)}
}

func advert(address string, reading byte) *AdvertisementEvent {
	return &AdvertisementEvent{
		Address: address,
		RSSI:    -60,
		Fields:  Fields{ManufacturerData: map[uint16][]byte{0xec88: {reading}}},
	}
}

func buttonAdvert() *AdvertisementEvent {
	return &AdvertisementEvent{Address: "C3:00:00:00:00:01", Fields: Fields{ManufacturerData: map[uint16][]byte{0x004c: {2, 21}}}}
}

func interested(t *testing.T, owner string, rules ...Rule) *Coalescer {
	t.Helper()

	c := NewCoalescer()
	if err := c.SetInterest(InterestEvent{Owner: owner, Rules: rules}); err != nil {
		t.Fatal(err)
	}

	return c
}

func expect(t *testing.T, c *Coalescer, a *AdvertisementEvent, want Verdict) {
	t.Helper()

	if got, _ := c.Admit(a); got != want {
		t.Errorf("%s %v: %v, want %v", a.Address, a.ManufacturerData, got, want)
	}
}

func TestCoalescer_NothingIsForwardedUntilSomebodyHasSaidWhatTheyWant(t *testing.T) {
	c := NewCoalescer()

	expect(t, c, advert("A4:C1:38:00:00:01", 1), NotWanted)

	if c.Interested() || c.ActiveScan() || len(c.Rules()) != 0 {
		t.Error("interested with nobody having said anything")
	}
}

func TestCoalescer_WhatMatchesNoRuleIsNotWanted(t *testing.T) {
	c := interested(t, "bluetooth", sensors())

	other := &AdvertisementEvent{Address: "A4:C1:38:00:00:01", Fields: Fields{ManufacturerData: map[uint16][]byte{0x004c: {1}}}}
	expect(t, c, other, NotWanted)
}

func TestCoalescer_TheFirstIsForwardedAtOnceAndIdenticalRepeatsAreNotUntilTheWindowHasPassed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := interested(t, "bluetooth", sensors())
		a := advert("A4:C1:38:00:00:01", 1)

		expect(t, c, a, Forward)

		// The same thing, over and over, louder or softer: not news.
		for range 5 {
			time.Sleep(5 * time.Second)

			again := advert("A4:C1:38:00:00:01", 1)
			again.RSSI = -90
			expect(t, c, again, Duplicate)
		}

		// 25 s so far. At 30 s the window has passed, so it is forwarded once more, to show the device is alive.
		time.Sleep(5 * time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)

		// And the window starts again from then.
		time.Sleep(10 * time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 1), Duplicate)
	})
}

func TestCoalescer_AChangedPayloadIsForwardedAtOnceWhateverTheWindow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := interested(t, "bluetooth", sensors())

		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
		expect(t, c, advert("A4:C1:38:00:00:01", 2), Forward)
		expect(t, c, advert("A4:C1:38:00:00:01", 2), Duplicate)

		// Going back to an earlier payload is a change too.
		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
	})
}

func TestCoalescer_EachAddressIsItsOwn(t *testing.T) {
	c := interested(t, "bluetooth", sensors())

	expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
	expect(t, c, advert("A4:C1:38:00:00:02", 1), Forward)
	expect(t, c, advert("A4:C1:38:00:00:01", 1), Duplicate)
	expect(t, c, advert("A4:C1:38:00:00:02", 1), Duplicate)
}

func TestCoalescer_ImmediateForwardsEveryAdvertisementEvenIdenticalOnes(t *testing.T) {
	c := interested(t, "bluetooth", button())

	for range 10 {
		expect(t, c, buttonAdvert(), Forward)
	}
}

func TestCoalescer_ARuleHasItsOwnWindow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		slow := sensors()
		slow.Window = 2 * time.Minute
		c := interested(t, "bluetooth", slow)

		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)

		time.Sleep(90 * time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 1), Duplicate)

		time.Sleep(30 * time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
	})
}

func TestCoalescer_WhenMoreThanOneRuleMatchesImmediateWinsAndOtherwiseTheShortestWindow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		long, short := sensors(), sensors()
		long.Window, short.Window = time.Minute, 10*time.Second

		c := interested(t, "bluetooth", long, short)

		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
		time.Sleep(11 * time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward) // the short window governs

		// A device matched by a Coalesce rule and by an Immediate one is Immediate.
		both := buttonAdvert()
		both.LocalName = "GVH"
		c = interested(t, "bluetooth", Rule{NamePrefix: "GVH"}, button())

		expect(t, c, both, Forward)
		expect(t, c, both, Forward)
	})
}

func TestCoalescer_TheRuleThatMatchedIsReturnedSoThatASourceCanCountByIt(t *testing.T) {
	c := interested(t, "bluetooth", sensors(), button())

	if _, rule := c.Admit(buttonAdvert()); rule == nil || rule.Name != "button" {
		t.Errorf("rule %+v", rule)
	}

	if _, rule := c.Admit(advert("A4:C1:38:00:00:01", 1)); rule == nil || rule.Name != "govee" {
		t.Errorf("rule %+v", rule)
	}

	if _, rule := c.Admit(&AdvertisementEvent{Address: "E0:00:00:00:00:00"}); rule != nil {
		t.Errorf("rule %+v for something not wanted", rule)
	}
}

func TestCoalescer_AnOwnersNewInterestReplacesItsOldOne(t *testing.T) {
	c := interested(t, "bluetooth", sensors())
	expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)

	// The sensors are no longer wanted; the button is.
	if err := c.SetInterest(InterestEvent{Owner: "bluetooth", Rules: []Rule{button()}}); err != nil {
		t.Fatal(err)
	}

	expect(t, c, advert("A4:C1:38:00:00:01", 2), NotWanted)
	expect(t, c, buttonAdvert(), Forward)
}

func TestCoalescer_AnEmptyInterestWithdrawsTheOwner(t *testing.T) {
	c := interested(t, "bluetooth", sensors())

	if err := c.SetInterest(InterestEvent{Owner: "bluetooth"}); err != nil {
		t.Fatal(err)
	}

	if c.Interested() {
		t.Error("still interested")
	}

	expect(t, c, advert("A4:C1:38:00:00:01", 1), NotWanted)
}

func TestCoalescer_OwnersAreUnitedAndEachCanBeWithdrawn(t *testing.T) {
	c := interested(t, "bluetooth", sensors())

	if err := c.SetInterest(InterestEvent{Owner: "doorbell", Rules: []Rule{button()}, ActiveScan: true}); err != nil {
		t.Fatal(err)
	}

	expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
	expect(t, c, buttonAdvert(), Forward)

	if len(c.Rules()) != 2 || !c.ActiveScan() {
		t.Errorf("rules %d, active %v", len(c.Rules()), c.ActiveScan())
	}

	_ = c.SetInterest(InterestEvent{Owner: "doorbell"})

	expect(t, c, buttonAdvert(), NotWanted)

	if c.ActiveScan() {
		t.Error("still scanning actively for an owner that left")
	}
}

func TestCoalescer_AnInterestNotRefreshedExpiresAfterThreeMinutes(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := interested(t, "bluetooth", sensors())

		time.Sleep(InterestTTL - time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)

		// Refreshed, so it lasts another three minutes.
		if err := c.SetInterest(InterestEvent{Owner: "bluetooth", Rules: []Rule{sensors()}}); err != nil {
			t.Fatal(err)
		}

		time.Sleep(InterestTTL - time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 2), Forward)

		time.Sleep(2 * time.Second)
		expect(t, c, advert("A4:C1:38:00:00:01", 3), NotWanted)

		if c.Interested() {
			t.Error("interested after it expired")
		}
	})
}

func TestCoalescer_AnInterestThatIsNotValidIsRefusedAndWhatWasThereStands(t *testing.T) {
	c := interested(t, "bluetooth", sensors())

	if err := c.SetInterest(InterestEvent{Owner: "bluetooth", Rules: []Rule{{}}}); err == nil {
		t.Fatal("accepted")
	}

	expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
}

func TestCoalescer_ItDoesNotKeepAnInterestThatTheCallerLaterChanges(t *testing.T) {
	rules := []Rule{sensors()}
	c := interested(t, "bluetooth", rules...)

	rules[0] = button()

	expect(t, c, advert("A4:C1:38:00:00:01", 1), Forward)
}

func TestCoalescer_WhatItRemembersIsBounded(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := interested(t, "bluetooth", sensors())

		// Many addresses, all recent: it starts again rather than growing.
		for i := 0; i < maxTracked+10; i++ {
			expect(t, c, advert(fmt.Sprintf("A4:C1:38:%02X:%02X:%02X", i>>16&0xff, i>>8&0xff, i&0xff), 1), Forward)
		}

		if len(c.last) > maxTracked {
			t.Errorf("remembers %d addresses", len(c.last))
		}

		// Old ones go first when it is full.
		time.Sleep(pruneAfter + time.Second)
		before := len(c.last)

		// (The interest has expired by now, as it does without being refreshed.)
		if err := c.SetInterest(InterestEvent{Owner: "bluetooth", Rules: []Rule{sensors()}}); err != nil {
			t.Fatal(err)
		}

		for i := 0; i < maxTracked-before+1; i++ {
			expect(t, c, advert(fmt.Sprintf("B4:C1:38:%02X:%02X:%02X", i>>16&0xff, i>>8&0xff, i&0xff), 1), Forward)
		}

		// The old ones were let go, and only the new ones are remembered.
		if want := maxTracked - before + 1; len(c.last) != want {
			t.Errorf("remembers %d addresses, want the %d new ones", len(c.last), want)
		}
	})
}

func TestCoalescer_IsSafeToUseFromSeveralGoroutines(t *testing.T) {
	c := NewCoalescer()
	done := make(chan struct{})

	go func() {
		defer close(done)

		for range 200 {
			_ = c.SetInterest(InterestEvent{Owner: "bluetooth", Rules: []Rule{sensors(), button()}})
			c.Interested()
			c.Rules()
		}
	}()

	for range 200 {
		c.Admit(advert("A4:C1:38:00:00:01", 1))
		c.Admit(buttonAdvert())
	}

	<-done
}

func TestVerdictAndLatencyNames(t *testing.T) {
	if NotWanted.String() != "not wanted" || Duplicate.String() != "duplicate" || Forward.String() != "forward" ||
		Immediate.String() != "immediate" || Coalesce.String() != "coalesce" ||
		ScanOff.String() != "off" || ScanPassive.String() != "passive" || ScanActive.String() != "active" {
		t.Error("names")
	}
}

func TestCoalescer_AdmitAsMatchesTheRulesAgainstWhatIsKnownOfTheDeviceAndDedupesTheAdvertisement(t *testing.T) {
	c := interested(t, "bluetooth", Rule{Name: "named", NamePrefix: "GVH"})

	// The advertisement alone says nothing the rule could match; the device is known to have the name.
	view := &AdvertisementEvent{Address: "A4:C1:38:00:00:01", Fields: Fields{LocalName: "GVH5075_0001"}}
	data := &AdvertisementEvent{Address: "A4:C1:38:00:00:01", Fields: Fields{ManufacturerData: map[uint16][]byte{0xec88: {1}}}}

	if verdict, rule := c.Admit(data); verdict != NotWanted || rule != nil {
		t.Errorf("without the view: %v %+v", verdict, rule)
	}

	if verdict, rule := c.AdmitAs(view, data); verdict != Forward || rule == nil || rule.Name != "named" {
		t.Errorf("with the view: %v %+v", verdict, rule)
	}

	// What was remembered is the advertisement, not the view.
	if verdict, _ := c.AdmitAs(view, data); verdict != Duplicate {
		t.Errorf("a repeat of the advertisement: %v", verdict)
	}

	other := &AdvertisementEvent{Address: "A4:C1:38:00:00:01", Fields: Fields{ManufacturerData: map[uint16][]byte{0xec88: {2}}}}
	if verdict, _ := c.AdmitAs(view, other); verdict != Forward {
		t.Errorf("a different advertisement: %v", verdict)
	}
}
