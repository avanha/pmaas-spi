package bttest_test

import (
	"testing"
	"time"

	"github.com/avanha/pmaas-spi/bluetooth"
	"github.com/avanha/pmaas-spi/bluetooth/bttest"
)

type echo struct{ n int }

func (e *echo) Hear(a *bluetooth.AdvertisementEvent, _ time.Time) bluetooth.Result {
	e.n++

	return bluetooth.Result{Battery: bluetooth.Percent(e.n)}
}

func TestAdvertAndKnown(t *testing.T) {
	// Flags, then a complete local name "sps" and a 16-bit service list of 0xfff0.
	a := bttest.Advert(t, "49-22-01-23-0a-1b", "020106"+"0409737073"+"0303f0ff")
	if a.Address != "49:22:01:23:0A:1B" || a.LocalName != "sps" || len(a.ServiceUUIDs) != 1 {
		t.Fatalf("%+v", a)
	}

	known := bttest.Known(t, "0303f0ff", "0409737073")
	if known.LocalName != "sps" || len(known.ServiceUUIDs) != 1 {
		t.Errorf("%+v", known)
	}
}

func TestFeed(t *testing.T) {
	results := bttest.Feed(&echo{}, bttest.Advert(t, "00:00:00:00:00:01", "020106"), bttest.Advert(t, "00:00:00:00:00:01", "020106"))
	if len(results) != 2 || *results[1].Battery != 2 {
		t.Errorf("%+v", results)
	}
}
