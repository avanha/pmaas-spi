package bluetooth

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/avanha/pmaas-spi/events"
)

// AddressType says what kind of address a device uses. Devices that matter here use public or static addresses;
// the many that rotate theirs (phones, trackers) are what the interest rules exist to drop.
type AddressType int

const (
	AddressUnknown AddressType = iota
	AddressPublic
	AddressRandom
)

func (t AddressType) String() string {
	switch t {
	case AddressPublic:
		return "public"
	case AddressRandom:
		return "random"
	default:
		return "unknown"
	}
}

// Fields is what an advertisement carried, decoded.
//
// A field that was not in the advertisement is nil or empty. That does NOT mean the device has no such thing: a name often
// comes in a scan response and the sensor data in the advertisement, so whoever wants the whole device keeps what it has
// seen before.
type Fields struct {
	LocalName string

	// ServiceUUIDs are 128-bit and lower case: "0000ec88-0000-1000-8000-00805f9b34fb".
	ServiceUUIDs []string

	// ManufacturerData is by company ID, without the ID itself.
	ManufacturerData map[uint16][]byte

	// ServiceData is by 128-bit UUID, in the same form as ServiceUUIDs, without the UUID itself.
	ServiceData map[string][]byte

	TxPower *int8
}

// AdvertisementEvent is broadcast by a source for each advertisement it hears and has been asked for (see InterestEvent).
// Events from one source are broadcast in the order they were heard.
type AdvertisementEvent struct {
	// Source is the entity ID of the source that heard it.
	Source string

	// Address is "A4:C1:38:CC:E2:F0": upper case, colon separated (see FormatAddress and NormalizeAddress).
	Address     string
	AddressType AddressType

	// RSSI is the signal strength in dBm, as this source heard it.
	RSSI int

	// Time is when the source received it, by the source's clock.
	Time time.Time

	Fields

	// Raw is the advertising data as received, when the source has it (a proxy does; BlueZ does not).
	Raw []byte
}

// PayloadKey identifies what the advertisement said, apart from who heard it and how loudly: two advertisements with the
// same key, from the same address, carried the same thing. It is the raw data if the source has it, and otherwise the
// fields.
func (a *AdvertisementEvent) PayloadKey() string {
	if len(a.Raw) > 0 {
		return "raw:" + string(a.Raw)
	}

	var b strings.Builder

	b.WriteString("fields:")
	b.WriteString(a.LocalName)
	b.WriteByte(0)

	uuids := append([]string(nil), a.ServiceUUIDs...)
	sort.Strings(uuids)
	b.WriteString(strings.Join(uuids, ","))
	b.WriteByte(0)

	ids := make([]int, 0, len(a.ManufacturerData))
	for id := range a.ManufacturerData {
		ids = append(ids, int(id))
	}

	sort.Ints(ids)

	for _, id := range ids {
		b.WriteString(strconv.Itoa(id))
		b.WriteByte('=')
		b.Write(a.ManufacturerData[uint16(id)])
		b.WriteByte(';')
	}

	b.WriteByte(0)

	keys := make([]string, 0, len(a.ServiceData))
	for uuid := range a.ServiceData {
		keys = append(keys, uuid)
	}

	sort.Strings(keys)

	for _, uuid := range keys {
		b.WriteString(uuid)
		b.WriteByte('=')
		b.Write(a.ServiceData[uuid])
		b.WriteByte(';')
	}

	b.WriteByte(0)

	if a.TxPower != nil {
		b.WriteString(strconv.Itoa(int(*a.TxPower)))
	}

	return b.String()
}

// Broadcaster is what a source needs to announce advertisements, which spi.IPMAASContainer provides.
type Broadcaster interface {
	BroadcastEvent(entityEventId string, event any) error
}

// PublishAdvertisement broadcasts the advertisement from its source.
func PublishAdvertisement(broadcaster Broadcaster, advertisement AdvertisementEvent) error {
	return broadcaster.BroadcastEvent(advertisement.Source, advertisement)
}

// AdvertisementPredicate is an events.EventPredicate matching AdvertisementEvents, for a consumer.
func AdvertisementPredicate(eventInfo *events.EventInfo) bool {
	_, ok := eventInfo.Event.(AdvertisementEvent)

	return ok
}

// IsTemporaryAddress is whether the address is one the device is going to change. A public address, and a random one that is
// static (its top two bits are 11), last for as long as the device does. A random address that is not static is private: its top
// bits are 01 for one that can be resolved with a key, 00 for one that can't, and phones, watches and trackers make a new one every
// quarter of an hour or so. A device with such an address is not worth remembering for long, or configuring by it. An address of
// an unknown type, or that is not an address, is not said to be temporary.
func IsTemporaryAddress(address string, addressType AddressType) bool {
	if addressType != AddressRandom {
		return false
	}

	normal, ok := NormalizeAddress(address)
	if !ok {
		return false
	}

	// The first byte is the most significant.
	var first byte

	for _, c := range normal[:2] {
		first <<= 4

		if c >= 'A' {
			first |= byte(c-'A') + 10
		} else {
			first |= byte(c - '0')
		}
	}

	return first&0xC0 != 0xC0
}

// FormatAddress is a 48-bit address, as proxies send it (most significant byte first in the number), in the form
// AdvertisementEvent uses.
func FormatAddress(address uint64) string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		address>>40&0xff, address>>32&0xff, address>>24&0xff, address>>16&0xff, address>>8&0xff, address&0xff)
}

// NormalizeAddress puts an address in the form AdvertisementEvent uses, from any case and with ":" or "-". It reports false
// if the text is not an address.
func NormalizeAddress(text string) (string, bool) {
	text = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(text), "-", ":"))

	parts := strings.Split(text, ":")
	if len(parts) != 6 {
		return "", false
	}

	for _, part := range parts {
		if len(part) != 2 || !isHex(part) {
			return "", false
		}
	}

	return text, true
}

func isHex(text string) bool {
	for _, c := range text {
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f') {
			return false
		}
	}

	return true
}

// baseUUIDSuffix is what follows the 16 or 32 bits in a short UUID's 128-bit form.
const baseUUIDSuffix = "-0000-1000-8000-00805f9b34fb"

// UUID16 is a 16-bit UUID in 128-bit form.
func UUID16(value uint16) string {
	return fmt.Sprintf("0000%04x%s", value, baseUUIDSuffix)
}

// UUID32 is a 32-bit UUID in 128-bit form.
func UUID32(value uint32) string {
	return fmt.Sprintf("%08x%s", value, baseUUIDSuffix)
}

// NormalizeUUID puts a UUID in the 128-bit lower case form events use, from the 4 digit, 8 digit or full form, in any case. It
// reports false if the text is not a UUID.
func NormalizeUUID(text string) (string, bool) {
	text = strings.ToLower(strings.TrimSpace(text))

	switch len(text) {
	case 4:
		if isHex(text) {
			return "0000" + text + baseUUIDSuffix, true
		}
	case 8:
		if isHex(text) {
			return text + baseUUIDSuffix, true
		}
	case 36:
		for i, c := range text {
			if i == 8 || i == 13 || i == 18 || i == 23 {
				if c != '-' {
					return "", false
				}

				continue
			}

			if !isHex(string(c)) {
				return "", false
			}
		}

		return text, true
	}

	return "", false
}
