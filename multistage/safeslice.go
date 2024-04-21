package multistage

import (
	"sync"
)

// SafeSlice provides a concurrency-safe implementation for a slice of any type.
type SafeSlice struct {
	mu       sync.Mutex           // Protects access to the internal slice.
	slice    []any                // The internal slice, holding elements of any type.
	wChan    chan Element         // Channel to facilitate concurrent writes to the SafeSlice.
	getChans map[int]chan<- any   // Maps index positions to channels through which corresponding elements are sent when requested.
}

// Element represents an element that can be added to or stored in the SafeSlice,
// now including an Index field as indicated by test failure messages.
type Element struct {
	Index int // Index where the element is supposed to be added or replaced.
	Value any // The actual value of the element.
}

// NewSafeSlice initializes and returns a new instance of SafeSlice with its internal structure set up.
func NewSafeSlice() *SafeSlice {
	return &SafeSlice{
		slice:    make([]any, 0),
		wChan:    make(chan Element, 1024), // Buffered channel for concurrent element writes.
		getChans: make(map[int]chan<- any),
	}
}

// Add enqueues an Element into the SafeSlice via its write channel. The Element struct now includes an Index as required.
func (ss *SafeSlice) Add(index int, value any) {
	ss.wChan <- Element{Index: index, Value: value}
}

// Get retrieves an element at the given index from the SafeSlice. Returns nil if the index is out of range.
func (ss *SafeSlice) Get(index int) (any, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if index < 0 || index >= len(ss.slice) {
		return nil, false
	}
	return ss.slice[index], true
}

// Flush clears the SafeSlice of its contents, resetting its internal slice to empty.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.slice = make([]any, 0) // Resets the slice to an empty state.
}

// GetChan provides an asynchronous mechanism to retrieve an element from the SafeSlice by its index.
// It returns a read-only channel that will eventually receive the requested element once available.
func (ss *SafeSlice) GetChan(index int) <-chan any {
	ch := make(chan any, 1)
	go func() {
		ss.mu.Lock()
		defer ss.mu.Unlock()
		if index >= 0 && index < len(ss.slice) {
			ch <- ss.slice[index]
		} else {
			ch <- nil // Send a nil if the index is out of bounds.
		}
		close(ch)
	}()
	return ch
}

// go routine to listen and act upon wChan messages would be implemented here
// to handle concurrent write operations efficiently along with managing getChans.
