package multistage

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	. "github.com/stevegt/goadapt"
)

func TestStructs(t *testing.T) {
	// Test the Element struct.
	e := Element{Index: 0, Value: 0}
	Tassert(t, e.Index == 0, "Element.Index is not 0")
	Tassert(t, e.Value == 0, "Element.Value is not 0")

	// Test the SafeSlice struct.
	ss := NewSafeSlice()
	Tassert(t, ss.getChans != nil, "SafeSlice.getChans is nil")
}

func TestSafeSliceOneThread(t *testing.T) {
	// testing in one thread
	ss := NewSafeSlice()

	// ensure Append returns the index of the element
	for i := 0; i < 10; i++ {
		index := ss.Append(i)
		Tassert(t, index == i, "Append returned %d, expected %d", index, i)
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
		// Append elements to the safeSlice
		for i := 0; i < 10; i++ {
			// delay to ensure the other goroutine is waiting
			time.Sleep(100 * time.Millisecond)
			ss.Append(i)
		}
	}()

	// Attempt to retrieve an element before any Appends.
	value, ok := ss.Get(0)
	Tassert(t, !ok, "Get returned true for index 0 before any Appends")
	Tassert(t, value == nil, "Get returned %v for index 0 before any Appends", value)

	// Retrieve channels using GetChan.  GetChan returns a channel
	// that contains the value when the value becomes available.
	chans := make([]<-chan any, 10)
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

func TestSafeSliceManyThreads(t *testing.T) {
	rand.Seed(1)
	ss := NewSafeSlice()

	size := 100
	expects := sync.Map{}
	wgAdd := sync.WaitGroup{}
	// Append elements to the safeSlice
	for i := 0; i < size; i++ {
		wgAdd.Add(1)
		go func(i int) {
			// wait a random amount of time before appending
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			index := ss.Append(i)
			Tassert(t, index >= 0, "Append returned %d, expected >= 0", index)
			expects.Store(index, i)
			value, ok := expects.Load(index)
			Tassert(t, ok, "Failed to store expected value")
			Tassert(t, value == i, "Expected value %v, got %v", i, value)
			// Pf("Appended %v at index %v\n", i, index)
			wgAdd.Done()
		}(i)
	}

	// Retrieve channels using GetChan.  GetChan returns a channel
	// that contains the value when the value becomes available.
	wgGet := sync.WaitGroup{}
	values := sync.Map{}
	for i := 0; i < size; i++ {
		wgGet.Add(1)
		go func(i int) {
			// wait a random amount of time before getting the channel
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			ch := ss.GetChan(i)
			// wait a random amount of time before retrieving the value
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			// get the value from the channel
			value := <-ch
			// Pf("Got %v from index %v\n", value, i)
			values.Store(i, value)
			wgGet.Done()
		}(i)
	}
	// wait until all the values are available
	wgGet.Wait()

	// wait until all the expected values are available
	wgAdd.Wait()

	// ensure the values are what we expect
	for i := 0; i < size; i++ {
		value, ok := values.Load(i)
		Tassert(t, ok, "Failed to get value")
		expect, ok := expects.Load(i)
		Tassert(t, ok, "Failed to get expected value")
		Tassert(t, value == expect, "GetChan got %v, expected %v", value, expect)
	}

}

func TestGetWaitOneThread(t *testing.T) {
	ss := NewSafeSlice()

	// Append elements to the safeSlice
	for i := 0; i < 10; i++ {
		ss.Append(i)
	}

	// GetWait should return the value at the index or ok == false
	// if the value does not become available within the timeout.
	// GetWait calls GetChan and waits for the value to be available
	// or the timeout to expire.
	for i := 0; i < 10; i++ {
		value, ok := ss.GetWait(i, 200*time.Millisecond)
		Tassert(t, ok, "GetWait returned false for index %d", i)
		Tassert(t, value == i, "GetWait returned %v for index %d", value, i)
	}

	// GetWait should return ok == false if the value does not become
	// available within the timeout.  In this case, the timeout is
	// 10ms.  We expect GetWait to return false for index 99 within
	// about 11ms.
	start := time.Now()
	_, ok := ss.GetWait(99, 10*time.Millisecond)
	stop := time.Now()
	Tassert(t, !ok, "GetWait returned true for index 99")
	Tassert(t, stop.Sub(start) > 10*time.Millisecond, "GetWait returned too quickly")
	Tassert(t, stop.Sub(start) < 20*time.Millisecond, "GetWait returned too slowly")

}

func TestGetWaitManyThreads(t *testing.T) {
	ss := NewSafeSlice()

	size := 100

	expects := sync.Map{}
	wgAdd := sync.WaitGroup{}
	// Append elements to the safeSlice
	for i := 0; i < size; i++ {
		wgAdd.Add(1)
		go func(i int) {
			// wait a random amount of time before appending
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			index := ss.Append(i)
			Tassert(t, index >= 0, "Append returned %d, expected >= 0", index)
			expects.Store(index, i)
			value, ok := expects.Load(index)
			Tassert(t, ok, "Failed to store expected value")
			Tassert(t, value == i, "Expected value %v, got %v", i, value)
			// Pf("Appended %v at index %v\n", i, index)
			wgAdd.Done()
		}(i)
	}

	// Retrieve channels using GetWait.  GetWait returns the value
	// when the value becomes available or the timeout expires.
	wgGet := sync.WaitGroup{}
	values := sync.Map{}
	for i := 0; i < size; i++ {
		wgGet.Add(1)
		go func(i int) {
			// wait a random amount of time before getting the value
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			value, ok := ss.GetWait(i, 200*time.Millisecond)
			Tassert(t, ok, "GetWait returned false for index %d", i)
			// Pf("Got %v from index %v\n", value, i)
			values.Store(i, value)
			wgGet.Done()
		}(i)
	}

	// wait until all the values are available
	wgGet.Wait()

	// wait until all the expected values are available
	wgAdd.Wait()

	// ensure the values are what we expect
	for i := 0; i < size; i++ {
		value, ok := values.Load(i)
		Tassert(t, ok, "Failed to get value")
		expect, ok := expects.Load(i)
		Tassert(t, ok, "Failed to get expected value")
		Tassert(t, value == expect, "GetWait got %v, expected %v", value, expect)
	}
}
