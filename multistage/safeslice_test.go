package multistage

import (
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
)

func TestSafeSliceOneThread(t *testing.T) {
	// testing in one thread
	ss := NewSafeSlice()

	// Add elements to the safeSlice
	for i := 0; i < 10; i++ {
		ss.Add(i)
	}

	// Retrieve elements.
	for i := 0; i < 10; i++ {
		value, ok := ss.Get(i)
		Tassert(t, ok, "Get returned false for index %d", i)
		Tassert(t, value == i, "Get returned %v for index %d", value, i)
	}

	// Test the Flush functionality.
	ss.Flush()
	if len(ss.slice) != 0 {
		t.Errorf("Flush did not empty the slice, remaining items: %v", ss.slice)
	}
}

func TestSafeSliceTwoThreads(t *testing.T) {

	ss := NewSafeSlice()

	go func() {
		// Add elements to the safeSlice
		for i := 0; i < 10; i++ {
			// delay to ensure the other goroutine is waiting
			time.Sleep(100 * time.Millisecond)
			ss.Add(i)
		}
	}()

	// Attempt to retrieve an element before any adds.
	value, ok := ss.Get(0)
	Tassert(t, !ok, "Get returned true for index 0 before any adds")
	Tassert(t, value == nil, "Get returned %v for index 0 before any adds", value)

	// Retrieve channels using GetChan.  GetChan returns a channel
	// that contains the value when the value becomes available.
	chans := make([]chan any, 10)
	for i := 0; i < 10; i++ {
		c := ss.GetChan(i)
		chans[i] = c
	}

	// Retrieve elements from the channels.
	for i := 0; i < 10; i++ {
		value := <-chans[i]
		Tassert(t, value == i, "GetChan returned %v for index %d", value, i)
	}
}
