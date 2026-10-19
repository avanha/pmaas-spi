// Package bttest is for testing a device model with advertisements captured as hex, as a proxy sends them.
package bttest

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/avanha/pmaas-spi/bluetooth"
)

// Start is when Feed begins.
var Start = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

// Hex is the bytes of the text, which may have spaces.
func Hex(t testing.TB, text string) []byte {
	t.Helper()

	raw, err := hex.DecodeString(strings.ReplaceAll(text, " ", ""))
	if err != nil {
		t.Fatalf("bad hex %q: %v", text, err)
	}

	return raw
}

// Advert is the advertisement from the address with the advertising data (hex), heard at -60 dBm from source "test".
func Advert(t testing.TB, address string, data string) *bluetooth.AdvertisementEvent {
	t.Helper()

	raw := bluetooth.TrimAdvertisingData(Hex(t, data))

	fields, err := bluetooth.ParseAdvertisingData(raw)
	if err != nil {
		t.Fatalf("bad advertising data %q: %v", data, err)
	}

	normal, ok := bluetooth.NormalizeAddress(address)
	if !ok {
		t.Fatalf("bad address %q", address)
	}

	return &bluetooth.AdvertisementEvent{Source: "test", Address: normal, RSSI: -60, Time: Start, Fields: fields, Raw: raw}
}

// Known is what a model is asked to Identify, from the advertising data of the device (hex): the fields of every advertisement
// merged, the way the plugin keeps them.
func Known(t testing.TB, data ...string) bluetooth.Fields {
	t.Helper()

	var known bluetooth.Fields

	for _, d := range data {
		fields := Advert(t, "00:00:00:00:00:00", d).Fields

		if fields.LocalName != "" {
			known.LocalName = fields.LocalName
		}

		for _, uuid := range fields.ServiceUUIDs {
			found := false

			for _, existing := range known.ServiceUUIDs {
				found = found || existing == uuid
			}

			if !found {
				known.ServiceUUIDs = append(known.ServiceUUIDs, uuid)
			}
		}
	}

	return known
}

// Feed gives the device the advertisements, a second apart from Start, and returns what it made of each.
func Feed(d bluetooth.Device, ads ...*bluetooth.AdvertisementEvent) []bluetooth.Result {
	return FeedEvery(d, time.Second, ads...)
}

// FeedEvery is Feed with the advertisements the step apart.
func FeedEvery(d bluetooth.Device, step time.Duration, ads ...*bluetooth.AdvertisementEvent) []bluetooth.Result {
	results := make([]bluetooth.Result, len(ads))
	now := Start

	for i, a := range ads {
		a.Time = now
		results[i] = d.Hear(a, now)
		now = now.Add(step)
	}

	return results
}
