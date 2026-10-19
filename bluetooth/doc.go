// Package bluetooth is how Bluetooth LE advertisements get from where they are heard to where they are understood.
//
// A source is a plugin that hears advertisements: the local radio through BlueZ, or an ESPHome Bluetooth proxy. It knows
// nothing about what the devices are. It broadcasts an AdvertisementEvent for each advertisement it has been asked for and
// registers an ISource entity so that it can be seen. A consumer, the bluetooth plugin, takes advertisements from any
// number of sources, works out what the devices are, and publishes them. It tells the sources what it wants to hear by
// broadcasting an InterestEvent, so that the many advertisements nobody wants are dropped where they are heard.
//
// The parts:
//
//   - AdvertisementEvent and Fields: what an advertisement carried. ParseAdvertisingData decodes the raw bytes a proxy
//     sends into Fields; BlueZ hands over decoded properties.
//   - Rule and InterestEvent: what a consumer wants, and how eagerly (Immediate, or Coalesce). A consumer broadcasts the
//     whole current set on start, when it changes, and every minute; see InterestEvent.
//   - Coalescer: what a source uses to apply the interest, so that every source treats an advertisement the same way.
//   - ISource and Source: the entity a source registers.
//   - DeviceModel, ConfiguredDevice and Device: what a consumer knows a kind of device by (Kind, Rules, Identify) and how a device
//     decodes its advertisements into a Result. A model lives in its own module (pmaas-bluetooth-devices has the common ones) and
//     the bluetooth plugin runs it. The bttest package tests one with advertisements captured as hex.
//
// Nothing here writes to Bluetooth. If something has to, the capability belongs on the source's entity as methods, as it
// does for a switch, and not on the event bus: a command has a target and an answer.
package bluetooth
