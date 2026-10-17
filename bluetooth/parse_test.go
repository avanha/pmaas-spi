package bluetooth

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"testing"
)

func unhex(t testing.TB, text string) []byte {
	t.Helper()

	b, err := hex.DecodeString(text)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

// Two advertisements captured from a Bluetooth proxy, from sensors that are ours.
const (
	goveeH5075 = "0d09475648353037355f45324630030388ec02010509ff88ec000339a664001aff4c000215494e54454c4c495f524f434b535f48575075f2ff0c"
	inkbirdSps = "0201060302f0ff04097370730aff6607970f0099606408"
)

func TestParseAdvertisingData_AGoveeH5075(t *testing.T) {
	fields, err := ParseAdvertisingData(unhex(t, goveeH5075))
	if err != nil {
		t.Fatal(err)
	}

	want := Fields{
		LocalName:    "GVH5075_E2F0",
		ServiceUUIDs: []string{"0000ec88-0000-1000-8000-00805f9b34fb"},
		ManufacturerData: map[uint16][]byte{
			0xec88: unhex(t, "000339a66400"),
			0x004c: unhex(t, "0215494e54454c4c495f524f434b535f48575075f2ff0c"),
		},
	}

	if !reflect.DeepEqual(fields, want) {
		t.Errorf("got  %+v\nwant %+v", fields, want)
	}
}

func TestParseAdvertisingData_AnInkbirdKeepsItsReadingInTheCompanyID(t *testing.T) {
	fields, err := ParseAdvertisingData(unhex(t, inkbirdSps))
	if err != nil {
		t.Fatal(err)
	}

	// The thermometer puts its reading where a company ID goes: 0x0766 and the bytes after it.
	want := Fields{
		LocalName:        "sps",
		ServiceUUIDs:     []string{"0000fff0-0000-1000-8000-00805f9b34fb"},
		ManufacturerData: map[uint16][]byte{0x0766: unhex(t, "970f00996064"+"08")},
	}

	if !reflect.DeepEqual(fields, want) {
		t.Errorf("got  %+v\nwant %+v", fields, want)
	}
}

func TestParseAdvertisingData_PaddingAfterAZeroLengthIsIgnored(t *testing.T) {
	// A proxy pads to a fixed size.
	raw := append(unhex(t, inkbirdSps), make([]byte, 39)...)

	fields, err := ParseAdvertisingData(raw)
	if err != nil || fields.LocalName != "sps" {
		t.Errorf("%+v, %v", fields, err)
	}
}

func TestParseAdvertisingData_Empty(t *testing.T) {
	for _, raw := range [][]byte{nil, {}, {0, 0, 0}} {
		fields, err := ParseAdvertisingData(raw)
		if err != nil || !reflect.DeepEqual(fields, Fields{}) {
			t.Errorf("%v: %+v, %v", raw, fields, err)
		}
	}
}

func TestParseAdvertisingData_AStructureThatRunsPastTheEndKeepsWhatCameBefore(t *testing.T) {
	// A name, then a structure that claims 9 bytes with 2 left.
	raw := unhex(t, "04097370730aff6607")

	fields, err := ParseAdvertisingData(raw)
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("error %v", err)
	}

	if fields.LocalName != "sps" || fields.ManufacturerData != nil {
		t.Errorf("fields %+v", fields)
	}
}

func TestParseAdvertisingData_ACompleteNameBeatsAShortenedOneInEitherOrder(t *testing.T) {
	for name, raw := range map[string]string{
		"short first":    "050841424344" + "0709414243445051",
		"complete first": "0709414243445051" + "050841424344",
	} {
		fields, err := ParseAdvertisingData(unhex(t, raw))
		if err != nil || fields.LocalName != "ABCDPQ" {
			t.Errorf("%s: %+v, %v", name, fields, err)
		}
	}

	fields, _ := ParseAdvertisingData(unhex(t, "050841424344"))
	if fields.LocalName != "ABCD" {
		t.Errorf("a shortened name alone: %q", fields.LocalName)
	}
}

func TestParseAdvertisingData_UUIDsOfEverySizeAreInTheLongForm(t *testing.T) {
	long := "00112233445566778899aabbccddeeff" // little endian on the air

	raw := unhex(t,
		"0303"+"88ec"+ // 16 bit list
			"0504"+"78563412"+ // 32 bit list, incomplete
			"1107"+long+ // 128 bit list
			"0303"+"88ec") // the same 16 bit UUID again

	fields, err := ParseAdvertisingData(raw)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"0000ec88-0000-1000-8000-00805f9b34fb",
		"12345678-0000-1000-8000-00805f9b34fb",
		"ffeeddcc-bbaa-9988-7766-554433221100",
	}

	if !reflect.DeepEqual(fields.ServiceUUIDs, want) {
		t.Errorf("got %v\nwant %v", fields.ServiceUUIDs, want)
	}
}

func TestParseAdvertisingData_ServiceDataOfEverySize(t *testing.T) {
	long := "00112233445566778899aabbccddeeff"

	raw := unhex(t,
		"0516"+"d2fc"+"0102"+ // 16 bit UUID 0xfcd2, data 01 02
			"0720"+"78563412"+"aa"+"bb"+ // 32 bit UUID, data aa bb
			"1321"+long+"cc"+"dd") // 128 bit UUID, data cc dd

	fields, err := ParseAdvertisingData(raw)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string][]byte{
		"0000fcd2-0000-1000-8000-00805f9b34fb": {1, 2},
		"12345678-0000-1000-8000-00805f9b34fb": {0xaa, 0xbb},
		"ffeeddcc-bbaa-9988-7766-554433221100": {0xcc, 0xdd},
	}

	if !reflect.DeepEqual(fields.ServiceData, want) {
		t.Errorf("got %v\nwant %v", fields.ServiceData, want)
	}
}

func TestParseAdvertisingData_TxPowerIsSigned(t *testing.T) {
	fields, _ := ParseAdvertisingData(unhex(t, "020af8"))

	if fields.TxPower == nil || *fields.TxPower != -8 {
		t.Errorf("tx power %v", fields.TxPower)
	}
}

func TestParseAdvertisingData_TooShortForTheirTypeAreSkipped(t *testing.T) {
	raw := unhex(t,
		"02ff"+"ab"+ // manufacturer data needs a 2 byte company ID
			"0216"+"d2"+ // service data needs a 2 byte UUID
			"0109"+ // a name of nothing
			"010a"+ // a transmit power with no power
			"0209"+"41") // and then a good name

	fields, err := ParseAdvertisingData(raw)
	if err != nil {
		t.Fatal(err)
	}

	if fields.ManufacturerData != nil || fields.ServiceData != nil || fields.TxPower != nil || fields.LocalName != "A" {
		t.Errorf("fields %+v", fields)
	}
}

func TestParseAdvertisingData_ACompanyIDAloneIsDataOfNothing(t *testing.T) {
	fields, _ := ParseAdvertisingData(unhex(t, "03ff4c00"))

	data, ok := fields.ManufacturerData[0x004c]
	if !ok || len(data) != 0 {
		t.Errorf("data %v, present %v", data, ok)
	}
}

func TestParseAdvertisingData_NothingRefersToTheInput(t *testing.T) {
	raw := unhex(t, goveeH5075)

	fields, _ := ParseAdvertisingData(raw)
	before := bytes.Clone(fields.ManufacturerData[0xec88])

	for i := range raw {
		raw[i] = 0xff
	}

	if !bytes.Equal(fields.ManufacturerData[0xec88], before) {
		t.Error("the fields changed when the input did")
	}
}

func TestParseAdvertisingData_ANameThatIsNotTextIsMadeIntoText(t *testing.T) {
	fields, _ := ParseAdvertisingData(unhex(t, "0409"+"41ff42"))

	if fields.LocalName != "A�B" {
		t.Errorf("name %q", fields.LocalName)
	}
}

func FuzzParseAdvertisingData(f *testing.F) {
	for _, seed := range []string{goveeH5075, inkbirdSps, "", "00", "ff", "04ff", "0516d2fc0102", "1107" + "00112233445566778899aabbccddeeff"} {
		raw, _ := hex.DecodeString(seed)
		f.Add(raw)
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		// Whatever the input, it must not panic, and what it returns must not point into the input.
		pristine := bytes.Clone(raw)

		fields, err := ParseAdvertisingData(raw)
		if err != nil && !errors.Is(err, ErrMalformed) {
			t.Fatalf("an error that is not ErrMalformed: %v", err)
		}

		for i := range raw {
			raw[i] ^= 0xff
		}

		again, _ := ParseAdvertisingData(pristine)
		if !reflect.DeepEqual(fields, again) {
			t.Fatalf("the result changed with the input:\n%+v\n%+v", fields, again)
		}
	})
}

func TestTrimAdvertisingData(t *testing.T) {
	good := unhex(t, inkbirdSps)

	for name, c := range map[string]struct {
		in   []byte
		want []byte
	}{
		"nothing":               {nil, nil},
		"all of it":             {good, good},
		"padded":                {append(bytes.Clone(good), make([]byte, 39)...), good},
		"all padding":           {make([]byte, 10), []byte{}},
		"a structure runs over": {unhex(t, "04097370730aff6607"), unhex(t, "0409737073")},
		"only a length":         {[]byte{5}, []byte{}},
	} {
		got := TrimAdvertisingData(c.in)
		if !bytes.Equal(got, c.want) || len(got) != len(c.want) {
			t.Errorf("%s: %x, want %x", name, got, c.want)
		}
	}
}
