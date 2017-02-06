package workers

import "github.com/currantlabs/gatt"

type Worker interface {
	OnDiscovery(p gatt.Peripheral, a *gatt.Advertisement, rssi int) bool
	OnConnect(p gatt.Peripheral, err error) error
}
