package main

import (
	"time"
)

type Weight struct {
	Time   time.Time `json:"datetime"`
	Weight float64   `json:"weight"`
}
