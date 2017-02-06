package main

import (
	"fmt"
	"log"

	"blegw/workers"

	"github.com/currantlabs/gatt"
	"github.com/currantlabs/gatt/examples/option"
)

type BleService struct {
	workers []*workers.Worker
	done    chan struct{}
}

// AddWorker - add new worker
func (s *BleService) AddWorker(worker workers.Worker) {
	if s.workers == nil {
		fmt.Println("Init workers.")
		s.workers = make([]*workers.Worker, 0, 0)
	}
	s.workers = append(s.workers, &worker)
}

func (s *BleService) Init() error {
	s.done = make(chan struct{})
	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return err
	}

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(s.onPeriphDiscovered),
		gatt.PeripheralConnected(s.onPeriphConnected),
		gatt.PeripheralDisconnected(s.onPeriphDisconnected),
	)

	d.Init(s.onStateChanged)
	<-s.done
	fmt.Println("Done")

	return nil
}

func (s *BleService) onStateChanged(d gatt.Device, state gatt.State) {
	fmt.Println("State:", state)
	switch state {
	case gatt.StatePoweredOn:
		fmt.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func (s *BleService) onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	isConnect := false
	for _, worker := range s.workers {
		ret := (*worker).OnDiscovery(p, a, rssi)
		if ret && !isConnect {
			isConnect = true
		}
	}

	if isConnect {
		p.Device().Connect(p)
	}
}

func (s *BleService) onPeriphConnected(p gatt.Peripheral, err error) {
	defer p.Device().CancelConnection(p)

	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}
	fmt.Printf("Set mtu done.\n")

	for _, worker := range s.workers {
		(*worker).OnConnect(p, err)
	}
}

func (s *BleService) onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	close(s.done)
}
