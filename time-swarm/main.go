package main

import (
	"math/rand"
	"time"

	. "github.com/stevegt/goadapt"
)

const (
	// gravity is the force that pulls all events toward the present time
	gravity = 0.1
)

// Event is a struct that represents a single event.
// The idea here is to create one goroutine per event and let them
// jostle for time slots.  Each event has a start and end time that it
// uses as an anchor and a priority that it uses as a spring-like
// force to pull it toward the anchor.  There is also a global
// gravity force that pulls all events toward the present time.

type Event struct {
	start    time.Time
	end      time.Time
	priority float64
}

// Run runs the event's goroutine
func (e *Event) Run() {
	go func() {
		time.Sleep(10 * time.Second)
	}()
}

func main() {
	// start a bunch of events
	for i := 0; i < 100000; i++ {
		event := &Event{
			start:    time.Now().Add(time.Duration(i) * time.Second),
			end:      time.Now().Add(time.Duration(i+1) * time.Second),
			priority: rand.Float64(),
		}
		event.Run()
	}
	Pf("")
	select {}
}
