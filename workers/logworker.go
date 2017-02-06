package workers

import (
	"fmt"

	"github.com/currantlabs/gatt"
)

type Log struct {
	Worker
}

func (s Log) OnDiscovery(p gatt.Peripheral, a *gatt.Advertisement, rssi int) bool {
	fmt.Printf("\nPeripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	fmt.Println("  Local Name        =", a.LocalName)
	fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	fmt.Println("  Service Data      =", a.ServiceData)
	fmt.Println("  RSSI      =", rssi)

	return false
}

func (s Log) OnConnect(p gatt.Peripheral, err error) error {
	ID := p.ID()
	fmt.Println("Connected to " + ID)

	return nil
}
