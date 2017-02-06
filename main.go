package main

import "blegw/workers"

func main() {
	service := BleService{}

	log := workers.Log{}
	service.AddWorker(log)

	scale := workers.Scale{}
	service.AddWorker(scale)

	service.Init()
}
