package main

import (
	"fmt"
	"math"
	"time"
)

func parseData(weightBytes []byte) *Weight {
	ctrlByte := weightBytes[0]

	isWeightRemoved := isBitSet(ctrlByte, 7)
	isStabilized := isBitSet(ctrlByte, 5)
	isLBSUnit := isBitSet(ctrlByte, 0)
	isCattyUnit := isBitSet(ctrlByte, 4)

	fmt.Printf("IsWeightRemoved: %v\n", isWeightRemoved)
	fmt.Printf("6 LSB Unknown: %v\n", isBitSet(ctrlByte, 6))
	fmt.Printf("IsStabilized: %v\n", isStabilized)
	fmt.Printf("IsCattyOrKg: %v\n", isCattyUnit)
	fmt.Printf("3 LSB Unknown: %v\n", isBitSet(ctrlByte, 3))
	fmt.Printf("2 LSB Unknown: %v\n", isBitSet(ctrlByte, 2))
	fmt.Printf("1 LSB Unknown: %v\n", isBitSet(ctrlByte, 1))
	fmt.Printf("IsLBS: %v\n", isLBSUnit)

	// // Only if the value is stabilized and the weight is *not* removed, the date is valid
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
