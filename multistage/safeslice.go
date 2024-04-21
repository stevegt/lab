package multistage

import (
	"sync"
)

// SafeSlice provides a concurrency-safe implementation for managing a dynamic slice.
// It ensures that all operations on the slice are safe to use concurrently through internal synchronization mechanisms.
type SafeSlice struct {
	mu       sync.RWMutex         // Protects access to the internal slice structure.
	slice    []any                // The underlying slice that stores the actual data.
	getChans map[int][]chan<- any // A map tracking channels waiting for data by index.
}

// Element represents a structure containing a value and its intended index within the SafeSlice.
// This allows for operations that target specific indices within the slice, facilitating insertions or updates.
type Element struct {
	Index int // The index in the slice where the value should be inserted or updated.
	Value any // The actual data value to be stored in the slice.
}

// NewSafeSlice initializes a new SafeSlice with the required internal structures.
func NewSafeSlice() *SafeSlice {
	ss := &SafeSlice{
		slice:    make([]any, 0),
		getChans: make(map[int][]chan<- any),
	}
	return ss
}

// Append adds a new element to the end of the SafeSlice.
func (ss *SafeSlice) Append(value any) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.slice = append(ss.slice, value)
	// Notify any waiting channels.
	index := len(ss.slice) - 1
	if chans, ok := ss.getChans[index]; ok {
		for _, ch := range chans {
			ch <- value
			close(ch)
		}
		delete(ss.getChans, index)
	}
}

// Get retrieves the element at the specified index from the SafeSlice.
// It returns the element along with a boolean indicating if the retrieval was successful.
func (ss *SafeSlice) Get(index int) (any, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	if index < 0 || index >= len(ss.slice) {
		return nil, false
	}
	return ss.slice[index], true
}

// Flush clears all elements from the SafeSlice, resetting its state.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.slice = make([]any, 0)
	ss.getChans = make(map[int][]chan<- any)
}

// GetChan returns a channel that can be used to retrieve the element at the specified index from the SafeSlice.
func (ss *SafeSlice) GetChan(index int) <-chan any {
	ch := make(chan any, 1) // Buffered channel for non-blocking send.
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if index < 0 {
		return nil
	}
	if index < len(ss.slice) {
		ch <- ss.slice[index]
	} else {
		ss.getChans[index] = append(ss.getChans[index], ch)
	}
	return ch
}
