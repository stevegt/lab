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

func TestDaemonThread(t *testing.T) {
	// Test to ensure the daemon thread is running and managing the
	// safeSlice.  The daemon thread should be responsible for all
	// slice operations and I/O.
	ss := NewSafeSlice()

	// get a channel that can be used to add elements to the safeSlice
	addChan := ss.AddChan()
	// ensure that ss.addChan is the same channel that is returned by
	// AddChan()
	Tassert(t, ss.addChan == addChan, "addChan is not the same as AddChan()")

	// add an element to the safeSlice using the channel
	element := Element{Index: 0, Value: 0}
	addChan <- element

	// ensure the daemon thread has added the element to the slice
	time.Sleep(100 * time.Millisecond)
	Tassert(t, len(ss.slice) == 1, "Daemon thread did not add element to slice")

	// check the value of the element in the slice
	value, ok := ss.Get(0)
	Tassert(t, ok, "Get returned false for index 0")
	Tassert(t, value == 0, "Get returned %v for index 0", value)

	// Add more elements to the safeSlice using the channel.
	// (This will replace the element at index 0.)
	for i := 0; i < 10; i++ {
		element := Element{Index: i, Value: i}
		addChan <- element
	}
	Tassert(t, len(ss.slice) == 10, "Expected 10 elements in slice, found %d", len(ss.slice))

	// Retrieve elements using GetChan.
	for i := 0; i < 10; i++ {
		c := ss.GetChan(i)
		value := <-c
		Tassert(t, value == i, "GetChan returned %v for index %d", value, i)
	}
}

/*
func TestNoMutex(t *testing.T) {
	// Test to ensure we're not using a mutex, but instead using a
	// daemon thread to manage the slice.  The daemon thread will
	// be responsible for all slice operations and I/O.
	ss := NewSafeSlice()

	// NewSafeSlice should have initialized the addChan
	// channel, the getChans slice, and started the daemon thread.

	// addChan is a channel that can be used to add elements to the safeSlice:
	//
	// addChan chan Element
	Tassert(t, ss.addChan != nil, "addChan is nil")

	// getChans is a map of slices of channels.  Each channel in each
	// getChans map entry is used to retrieve a single element from
	// the safeSlice:
	//
	// getChans map[int][]chan any
	Tassert(t, ss.getChans != nil, "getChan is nil")

	// Add elements to the safeSlice using the addChan.  Normally,
	// Add would do this, but we're going to test the addChan directly.
	for i := 0; i < 10; i++ {
		element := Element{Index: i, Value: i}
		ss.addChan <- element
	}

	// Populate the getChans map entries.  Normally, GetChan would do
	// this, but we're going to test the getChans map directly.
	gc := make(map[int]chan any)
	for i := 0; i < 10; i++ {
		_, ok := ss.getChans[i]
		Tassert(t, !ok, "getChans[%d] is already initialized", i)
		getChan := make(chan any)
		ss.getChans[i] = append(ss.getChans[i], getChan)
		gc[i] = getChan
	}

	// Retrieve elements using the getChans map.
	for i := 0; i < 10; i++ {
		value := <-gc[i]
		Tassert(t, value == i, "GetChan returned %v for index %d", value, i)
	}
}
*/
