package multistage

import (
	"sync"
)

// Element is a struct that holds an index and a value. This struct is used
// with the AddChan method to allow concurrent addition of elements at specified
// indices in the SafeSlice. However, for the current implementation, the index is ignored,
// and elements are appended in the sequence they are received.
type Element struct {
	Index int // Index where the value should be added - not used in the current implementation
	Value any // The value to be added
}

// SafeSlice is a thread-safe data structure that supports concurrent read, write, and notification operations.
type SafeSlice struct {
	mu       sync.Mutex
	slice    []any
	getChans map[int][]chan any
	addChan  chan Element
}

// NewSafeSlice creates a new SafeSlice instance.
func NewSafeSlice() *SafeSlice {
	ss := &SafeSlice{
		getChans: make(map[int][]chan any),
		addChan:  make(chan Element),
	}
	return ss
}

// Add appends an item to the end of the slice in a thread-safe manner
// and notifies any goroutines waiting for this particular index.
func (ss *SafeSlice) Add(value any) {
	ss.mu.Lock()
	index := len(ss.slice)
	ss.slice = append(ss.slice, value)

	// Notify all waiting goroutines for this index.
	if waitingChans, ok := ss.getChans[index]; ok {
		for _, ch := range waitingChans {
			ch <- value // Send the value to the waiting goroutine.
			close(ch)   // Close the channel to signify the value has been sent.
		}
		delete(ss.getChans, index) // Remove the getChans for this index as they have been notified.
	}

	ss.mu.Unlock()
}

// Get retrieves an item from the slice by index in a thread-safe manner.
// Returns the item and true if the index is within bounds; otherwise, returns nil and false.
func (ss *SafeSlice) Get(index int) (value any, ok bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if index < 0 || index >= len(ss.slice) {
		return nil, false
	}

	return ss.slice[index], true
}

// Flush clears the slice and its associated state in a thread-safe manner.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	ss.slice = nil
	ss.getChans = make(map[int][]chan any)
	ss.mu.Unlock()
}

// GetChan returns a channel that will either immediately receive the
// requested item (if it is already present in the slice) or will receive
// the item once it is added to the slice at the specified index.
func (ss *SafeSlice) GetChan(index int) chan any {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ch := make(chan any, 1)

	// If the requested index is already within the bounds of the slice,
	// send the item immediately.
	if index >= 0 && index < len(ss.slice) {
		ch <- ss.slice[index]
		close(ch)
		return ch
	}

	// If the item at the requested index is not yet present, add the channel
	// to the list of getChans to be notified when the item is added.
	ss.getChans[index] = append(ss.getChans[index], ch)
	return ch
}
