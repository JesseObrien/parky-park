package main

import (
	"fmt"
	"log"
	"time"
)

var baseRate = 300

// Ticket for a customer
type Ticket struct {
	ID       int64
	TimeIn   time.Time
	TimePaid time.Time
	Paid     int64
	Card     string
}

// Increase the ticket rate per iteration
func increaseRate(owing int, times int) int {
	for i := 1; i <= times; i++ {
		increase := owing / 2
		owing += increase
	}

	return owing
}

// CalculateOwing will give back a representation of the owing amount
func (t *Ticket) CalculateOwing() int {
	hours := time.Since(t.TimeIn).Hours()

	if hours > 1.0 && hours < 3.0 {
		return increaseRate(baseRate, 1)
	}

	if hours >= 3.0 && hours <= 6.0 {
		return increaseRate(baseRate, 2)
	}

	if hours > 6.0 {
		return increaseRate(baseRate, 3)
	}

	return baseRate
}

// ShowOwing will show the user what's owing in a nice format
func (t *Ticket) ShowOwing() string {
	log.Println(float64(t.CalculateOwing()) / 100)
	return fmt.Sprintf("$%.2f", float64(t.CalculateOwing())/100)
}
