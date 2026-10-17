package bluetooth

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ErrMalformed is returned by ParseAdvertisingData when a structure runs past the end of the data. The fields decoded
// before it are still returned.
var ErrMalformed = errors.New("malformed advertising data")

// Advertising data types (the Bluetooth assigned numbers for "AD types") that ParseAdvertisingData understands. Others
// are skipped.
const (
	adIncomplete16   = 0x02
	adComplete16     = 0x03
	adIncomplete32   = 0x04
	adComplete32     = 0x05
	adIncomplete128  = 0x06
	adComplete128    = 0x07
	adShortName      = 0x08
	adCompleteName   = 0x09
	adTxPower        = 0x0a
	adServiceData16  = 0x16
	adServiceData32  = 0x20
	adServiceData128 = 0x21
	adManufacturer   = 0xff
)

// TrimAdvertisingData is the advertising data without the padding a proxy adds to fill a fixed size: everything up to the
// first structure of zero length, or up to the last whole structure if one runs past the end. It does not copy.
func TrimAdvertisingData(raw []byte) []byte {
	i := 0

	for i < len(raw) && raw[i] != 0 {
		next := i + 1 + int(raw[i])
		if next > len(raw) {
			break
		}

		i = next
	}

	return raw[:i]
}

// ParseAdvertisingData decodes the advertising data of one advertisement, as a proxy sends it: structures of a length byte,
// a type byte and the rest, one after another. A length of zero ends the data (a proxy pads it to a fixed size).
//
// Nothing in the result refers to raw, so it can be reused. If a structure claims more bytes than are left, decoding stops
// with what was found before it and an error wrapping ErrMalformed. A structure of a type that isn't understood, or too short
// for its type, is skipped.
func ParseAdvertisingData(raw []byte) (Fields, error) {
	var fields Fields

	var shortName string

	for i := 0; i < len(raw); {
		length := int(raw[i])
		if length == 0 {
			break
		}

		if i+1+length > len(raw) {
			fields.finishName(shortName)

			return fields, fmt.Errorf("%w: a structure at byte %d claims %d bytes and %d remain", ErrMalformed, i, length, len(raw)-i-1)
		}

		kind, value := raw[i+1], raw[i+2:i+1+length]
		i += 1 + length

		switch kind {
		case adIncomplete16, adComplete16:
			fields.addUUIDs(value, 2)
		case adIncomplete32, adComplete32:
			fields.addUUIDs(value, 4)
		case adIncomplete128, adComplete128:
			fields.addUUIDs(value, 16)
		case adShortName:
			shortName = name(value)
		case adCompleteName:
			fields.LocalName = name(value)
		case adTxPower:
			if len(value) >= 1 {
				power := int8(value[0])
				fields.TxPower = &power
			}
		case adManufacturer:
			if len(value) >= 2 {
				if fields.ManufacturerData == nil {
					fields.ManufacturerData = make(map[uint16][]byte)
				}

				fields.ManufacturerData[uint16(value[0])|uint16(value[1])<<8] = slices.Clone(value[2:])
			}
		case adServiceData16:
			fields.addServiceData(value, 2)
		case adServiceData32:
			fields.addServiceData(value, 4)
		case adServiceData128:
			fields.addServiceData(value, 16)
		}
	}

	fields.finishName(shortName)

	return fields, nil
}

// finishName uses the shortened name if there was no complete one.
func (f *Fields) finishName(shortName string) {
	if f.LocalName == "" {
		f.LocalName = shortName
	}
}

func name(value []byte) string {
	return strings.ToValidUTF8(string(value), "�")
}

// addUUIDs adds a list of UUIDs of the given size in bytes, each little endian.
func (f *Fields) addUUIDs(value []byte, size int) {
	for ; len(value) >= size; value = value[size:] {
		uuid := uuidOf(value[:size])
		if !slices.Contains(f.ServiceUUIDs, uuid) {
			f.ServiceUUIDs = append(f.ServiceUUIDs, uuid)
		}
	}
}

// addServiceData adds service data whose UUID, of the given size in bytes and little endian, comes first.
func (f *Fields) addServiceData(value []byte, size int) {
	if len(value) < size {
		return
	}

	if f.ServiceData == nil {
		f.ServiceData = make(map[string][]byte)
	}

	f.ServiceData[uuidOf(value[:size])] = slices.Clone(value[size:])
}

// uuidOf is a UUID of 2, 4 or 16 bytes, little endian as advertised, in the 128-bit form events use.
func uuidOf(little []byte) string {
	switch len(little) {
	case 2:
		return UUID16(uint16(little[0]) | uint16(little[1])<<8)
	case 4:
		return UUID32(uint32(little[0]) | uint32(little[1])<<8 | uint32(little[2])<<16 | uint32(little[3])<<24)
	}

	big := make([]byte, 16)
	for i, b := range little {
		big[15-i] = b
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", big[0:4], big[4:6], big[6:8], big[8:10], big[10:16])
}
