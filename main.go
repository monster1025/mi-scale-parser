package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/currantlabs/gatt"
	"github.com/currantlabs/gatt/examples/option"
)

var done = make(chan struct{})
var weights []*Weight
var isdone bool

func onStateChanged(d gatt.Device, s gatt.State) {
	fmt.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		fmt.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	ID := p.ID()
	//xiaomi scale
	if !strings.HasPrefix(ID, "88:0F:10") {
		return
	}

	fmt.Printf("\nPeripheral ID:%s, NAME:(%s)\n", ID, p.Name())

	fmt.Printf("%v+", a)
	fmt.Println("  Local Name        =", a.LocalName)
	fmt.Println("  TX Power Level    =", a.TxPowerLevel)
	fmt.Println("  Manufacturer Data =", a.ManufacturerData)
	fmt.Println("  Service Data      =", a.ServiceData)
	fmt.Println("  RSSI      =", rssi)

	p.Device().StopScanning()
	var mu sync.Mutex
	count := 0
	mu.Lock()
	go func(count int) {
		time.Sleep(time.Duration(count) * 5 * time.Second)
		fmt.Println("Connecting...")
		p.Device().Connect(p)
	}(count)
	count++
	mu.Unlock()
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	var WEIGHT_MEASUREMENT_SERVICE = gatt.MustParseUUID("181d")
	var WEIGHT_MEASUREMENT_HISTORY_CHARACTERISTIC = gatt.MustParseUUID("00002a2f-0000-3512-2118-0009af100700")
	// var WEIGHT_MEASUREMENT_TIME_CHARACTERISTIC = gatt.MustParseUUID("2a2b")
	var WEIGHT_MEASUREMENT_CONFIG = gatt.MustParseUUID("2902")

	fmt.Println("Connected")
	defer p.Device().CancelConnection(p)
	if err := p.SetMTU(500); err != nil {
		fmt.Printf("Failed to set MTU, err: %s\n", err)
	}
	fmt.Printf("Set mtu done.\n")

	// Discovery services
	services, err := p.DiscoverServices(nil)
	fmt.Printf("Services: %v+", services)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}
	service := getServiceByUUID(WEIGHT_MEASUREMENT_SERVICE, services)
	fmt.Printf("Service Found %s\n", service.Name())
	chars, _ := p.DiscoverCharacteristics(nil, service)

	// char := getCharByUUID(WEIGHT_MEASUREMENT_TIME_CHARACTERISTIC, chars)
	// b, err := p.ReadCharacteristic(char)
	// if err != nil {
	// 	fmt.Printf("Failed to read characteristic, err: %s\n", err)
	// 	return
	// }
	// parseData(b)
	// return
	// fmt.Printf("    value         %x | %q\n", b, b)

	histChar := getCharByUUID(WEIGHT_MEASUREMENT_HISTORY_CHARACTERISTIC, chars)
	descriptors, _ := p.DiscoverDescriptors(nil, histChar)
	descriptor := getDescriptorByUUID(WEIGHT_MEASUREMENT_CONFIG, descriptors)
	p.WriteDescriptor(descriptor, []byte{0x02, 0x00})

	isdone = false
	p.SetNotifyValue(histChar, func(c *gatt.Characteristic, b []byte, e error) {
		log.Printf("Got back from %s: %#x\n", c.UUID(), b)
		//weight :=
		if len(b) == 20 {
			weight := parseData(b[0:10])
			if weight != nil {
				weights = append(weights, weight)
			}
			weight = parseData(b[10:20])
			if weight != nil {
				weights = append(weights, weight)
			}
		} else if len(b) == 10 {
			weight := parseData(b)
			if weight != nil {
				weights = append(weights, weight)
			}
		} else if (len(b) == 1) && b[0] == 0x3 {
			isdone = true
			return
		} else {
			fmt.Printf("Invalid data: %v", b)
		}
	})
	p.WriteCharacteristic(histChar, []byte{0x02}, true)

	fmt.Printf("Waiting for 120 seconds to get some notifiations, if any.\n")
	for i := 0; i < 120; i++ {
		if isdone {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func getDescriptorByUUID(uuid gatt.UUID, descriptors []*gatt.Descriptor) *gatt.Descriptor {
	for _, descriptor := range descriptors {
		if !descriptor.UUID().Equal(uuid) {
			continue
		}
		return descriptor
	}
	return nil
}

func getServiceByUUID(uuid gatt.UUID, services []*gatt.Service) *gatt.Service {
	for _, service := range services {
		if !service.UUID().Equal(uuid) {
			continue
		}
		return service
	}
	return nil
}

func getCharByUUID(uid gatt.UUID, chars []*gatt.Characteristic) *gatt.Characteristic {
	for _, c := range chars {
		if !c.UUID().Equal(uid) {
			continue
		}
		return c
	}
	return nil
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	fmt.Println("Disconnected")
	close(done)
}

func main() {
	d, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	// Register handlers.
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	<-done
	if isdone {
		writeToFile()
	}
}

func writeToFile() {
	jsonbytes, err := json.Marshal(weights)
	if err != nil {
		panic(err)
	}
	var out bytes.Buffer
	err = json.Indent(&out, jsonbytes, "", "\t")
	ioutil.WriteFile("weights.json", out.Bytes(), 0644)
}
