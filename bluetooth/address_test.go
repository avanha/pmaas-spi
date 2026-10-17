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

func TestIsTemporaryAddress(t *testing.T) {
	for _, c := range []struct {
		address string
		kind    AddressType
		want    bool
		why     string
	}{
		{"A4:C1:38:CC:E2:F0", AddressPublic, false, "a public address, a Govee"},
		{"4C:00:00:00:00:00", AddressPublic, false, "a public address whose top bits would read as private"},
		{"C3:00:00:00:00:01", AddressRandom, false, "static random: 11"},
		{"FF:FF:FF:FF:FF:FF", AddressRandom, false, "static random: 11"},
		{"F3:8C:41:E1:0B:E1", AddressRandom, false, "static random seen on a proxy"},
		{"44:DC:FA:B6:5B:00", AddressRandom, true, "resolvable private: 01, an Apple device"},
		{"7F:7E:AD:F5:7E:EC", AddressRandom, true, "resolvable private: 01"},
		{"4B:41:7B:59:82:FA", AddressRandom, true, "resolvable private: 01"},
		{"0F:51:C0:6F:70:B3", AddressRandom, true, "non-resolvable private: 00"},
		{"25:71:9C:AE:14:6D", AddressRandom, true, "non-resolvable private: 00"},
		{"BF:00:00:00:00:00", AddressRandom, true, "10 is not a valid random address, and not a static one"},
		{"44:DC:FA:B6:5B:00", AddressUnknown, false, "a type not known is not said to be temporary"},
		{"nonsense", AddressRandom, false, "not an address"},
		{"c3:00:00:00:00:01", AddressRandom, false, "lower case is the same address"},
	} {
		if got := IsTemporaryAddress(c.address, c.kind); got != c.want {
			t.Errorf("%s %v (%s): %v, want %v", c.address, c.kind, c.why, got, c.want)
		}
	}
}
