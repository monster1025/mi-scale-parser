package main

import "blegw/workers"

func main() {
	service := BleService{}

	scale := workers.Scale{}
	service.AddWorker(scale)

	service.Init()
}
