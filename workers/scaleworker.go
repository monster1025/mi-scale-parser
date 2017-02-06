package workers

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/currantlabs/gatt"
)

type Weight struct {
	Time   time.Time
	Weight float64
}

type Scale struct {
	Worker
}

var WEIGHT_MEASUREMENT_SERVICE = gatt.MustParseUUID("181d")
var WEIGHT_MEASUREMENT_HISTORY_CHARACTERISTIC = gatt.MustParseUUID("00002a2f-0000-3512-2118-0009af100700")
var WEIGHT_MEASUREMENT_TIME_CHARACTERISTIC = gatt.MustParseUUID("2a2b")
var WEIGHT_MEASUREMENT_CONFIG = gatt.MustParseUUID("2902")

func (s Scale) OnDiscovery(p gatt.Peripheral, a *gatt.Advertisement, rssi int) bool {
	ID := p.ID()
	if !strings.HasPrefix(ID, "88:0F:10") {
		return false
	}

	return true
}

func (s Scale) OnConnect(p gatt.Peripheral, err error) error {
	ID := p.ID()
	//not our device
	if !strings.HasPrefix(ID, "88:0F:10") {
		return nil
	}

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
		return nil
	}

	service := getServiceByUUID(WEIGHT_MEASUREMENT_SERVICE, services)
	fmt.Printf("Service Found %s\n", service.Name())
	chars, _ := p.DiscoverCharacteristics(nil, service)

	char := getCharByUUID(WEIGHT_MEASUREMENT_TIME_CHARACTERISTIC, chars)
	b, err := p.ReadCharacteristic(char)
	if err != nil {
		fmt.Printf("Failed to read characteristic, err: %s\n", err)
		return nil
	}
	fmt.Printf("    value         %x | %q\n", b, b)

	histChar := getCharByUUID(WEIGHT_MEASUREMENT_HISTORY_CHARACTERISTIC, chars)
	descriptors, _ := p.DiscoverDescriptors(nil, histChar)
	descriptor := getDescriptorByUUID(WEIGHT_MEASUREMENT_CONFIG, descriptors)
	p.WriteDescriptor(descriptor, []byte{0x02, 0x00})

	p.SetNotifyValue(histChar, func(c *gatt.Characteristic, b []byte, e error) {
		log.Printf("Got back from %s: %#x\n", c.UUID(), b)
		//weight :=
		if len(b) == 20 {
			first := b[0:10]
			parseData(first)
			second := b[10:20]
			parseData(second)
		} else if len(b) == 10 {
			parseData(b)
		}
	})
	p.WriteCharacteristic(histChar, []byte{0x02}, true)

	fmt.Printf("Waiting for 5 seconds to get some notifiations, if any.\n")
	time.Sleep(10 * time.Second)

	return nil
}

func parseData(weightBytes []byte) *Weight {
	ctrlByte := weightBytes[0]

	isWeightRemoved := isBitSet(ctrlByte, 7)
	isStabilized := isBitSet(ctrlByte, 5)
	isLBSUnit := isBitSet(ctrlByte, 0)
	isCattyUnit := isBitSet(ctrlByte, 4)

	// Only if the value is stabilized and the weight is *not* removed, the date is valid
	if isStabilized && !isWeightRemoved {
		wb4 := int(weightBytes[4])
		wb3 := int(weightBytes[3])
		year := (wb4<<8 | wb3)
		month := int(weightBytes[5])
		day := int(weightBytes[6])
		hours := int(weightBytes[7])
		min := int(weightBytes[8])
		sec := int(weightBytes[9])

		// Is the year plausible? Check if the year is in the range of 20 years...
		now := time.Now()
		if math.Abs(float64(now.Year()-year)) > 20 || year > now.Year() {
			return nil
		}

		wb1 := int(weightBytes[1])
		wb2 := int(weightBytes[2])
		weightVal := 0.0
		if isLBSUnit || isCattyUnit {
			weightVal = float64(((wb2)<<8)|(wb1)) / 100.0
		} else {
			weightVal = float64(((wb2)<<8)|(wb1)) / 200.0
		}

		t := time.Date(year, time.Month(month), day, hours, min, sec, 0, time.Local)
		weight := &Weight{Weight: weightVal, Time: t}

		// fmt.Printf("Payload is: %#x\n", weightBytes)
		fmt.Printf("[+] Year: %v, Month: %v, day: %v, hr: %v, min: %v, sec: %v = %v \n", year, month, day, hours, min, sec, weightVal)
		return weight
	}
	return nil
}

func isBitSet(value byte, bit uint) bool {
	return (value & (1 << bit)) != 0
}
