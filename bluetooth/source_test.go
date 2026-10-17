package bluetooth

import (
	"sync"
	"testing"
	"time"

	"github.com/avanha/pmaas-spi/environment"
)

func TestSourceState_TracksConnectivityScanningAndCounts(t *testing.T) {
	s := NewSourceState("Local Bluetooth", "bluez")

	if got := s.GetSourceData(); got.Name != "Local Bluetooth" || got.Kind != "bluez" || got.Connectivity != environment.ConnectivityUnknown {
		t.Fatalf("a new source: %+v", got)
	}

	at := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

	s.SetConnectivity(environment.ConnectivityOnline, at)
	s.SetConnectivity(environment.ConnectivityOnline, at.Add(time.Hour)) // no change: the time stands
	s.SetScanning(ScanActive)
	s.Heard(at)
	s.Heard(at.Add(time.Second))
	s.Forwarded()
	s.Sighted()
	s.Sighted()
	s.Dropped()
	s.Dropped()

	got := s.GetSourceData()
	if got.Connectivity != environment.ConnectivityOnline || !got.OnlineSince.Equal(at) || got.Scanning != ScanActive ||
		got.Heard != 2 || !got.LastHeard.Equal(at.Add(time.Second)) || got.Forwarded != 1 || got.Sighted != 2 || got.Dropped != 2 {
		t.Errorf("%+v", got)
	}

	s.SetConnectivity(environment.ConnectivityOffline, at.Add(time.Minute))
	s.SetScanning(ScanOff)

	got = s.GetSourceData()
	if got.Connectivity != environment.ConnectivityOffline || !got.OfflineSince.Equal(at.Add(time.Minute)) || !got.OnlineSince.Equal(at) || got.Scanning != ScanOff {
		t.Errorf("%+v", got)
	}

	// Unknown moves no time.
	s.SetConnectivity(environment.ConnectivityUnknown, at.Add(time.Hour))

	if got := s.GetSourceData(); !got.OfflineSince.Equal(at.Add(time.Minute)) || !got.OnlineSince.Equal(at) {
		t.Errorf("%+v", got)
	}
}

func TestSourceState_IsSafeToUseFromSeveralGoroutines(t *testing.T) {
	s := NewSourceState("x", "y")

	var wg sync.WaitGroup

	for range 4 {
		wg.Go(func() {
			for range 200 {
				s.Heard(time.Now())
				s.Forwarded()
				s.SetScanning(ScanPassive)
				s.GetSourceData()
			}
		})
	}

	wg.Wait()

	if got := s.GetSourceData(); got.Heard != 800 || got.Forwarded != 800 {
		t.Errorf("%+v", got)
	}
}
