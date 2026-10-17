package bluetooth

import "testing"

func TestFormatAddress(t *testing.T) {
	if got := FormatAddress(0xA4C138CCE2F0); got != "A4:C1:38:CC:E2:F0" {
		t.Errorf("got %q", got)
	}

	if got := FormatAddress(0x0000000000AB); got != "00:00:00:00:00:AB" {
		t.Errorf("got %q", got)
	}
}

func TestNormalizeAddress(t *testing.T) {
	for in, want := range map[string]string{
		"A4:C1:38:CC:E2:F0":    "A4:C1:38:CC:E2:F0",
		"a4:c1:38:cc:e2:f0":    "A4:C1:38:CC:E2:F0",
		"a4-c1-38-cc-e2-f0":    "A4:C1:38:CC:E2:F0",
		"  a4:c1:38:cc:e2:f0 ": "A4:C1:38:CC:E2:F0",
	} {
		if got, ok := NormalizeAddress(in); !ok || got != want {
			t.Errorf("%q: %q, %v", in, got, ok)
		}
	}

	for _, bad := range []string{"", "A4:C1:38:CC:E2", "A4:C1:38:CC:E2:F0:00", "A4:C1:38:CC:E2:GG", "A4C138CCE2F0", "A4:C1:38:CC:E2:F"} {
		if got, ok := NormalizeAddress(bad); ok {
			t.Errorf("%q accepted as %q", bad, got)
		}
	}
}

func TestNormalizeUUID(t *testing.T) {
	for in, want := range map[string]string{
		"ec88":                                 "0000ec88-0000-1000-8000-00805f9b34fb",
		"EC88":                                 "0000ec88-0000-1000-8000-00805f9b34fb",
		"12345678":                             "12345678-0000-1000-8000-00805f9b34fb",
		"0000EC88-0000-1000-8000-00805F9B34FB": "0000ec88-0000-1000-8000-00805f9b34fb",
	} {
		if got, ok := NormalizeUUID(in); !ok || got != want {
			t.Errorf("%q: %q, %v", in, got, ok)
		}
	}

	for _, bad := range []string{"", "ec8", "ec88x", "0000ec88000010008000-00805f9b34fb-xx", "0000ec88_0000-1000-8000-00805f9b34fb", "0000ec88-0000-1000-8000-00805f9b34fz"} {
		if got, ok := NormalizeUUID(bad); ok {
			t.Errorf("%q accepted as %q", bad, got)
		}
	}
}

func TestAddressType_String(t *testing.T) {
	if AddressPublic.String() != "public" || AddressRandom.String() != "random" || AddressUnknown.String() != "unknown" {
		t.Error("names")
	}
}
