package multistage

import (
	"sync"
)

// SafeSlice is a thread-safe slice supporting concurrent operations.
type SafeSlice struct {
	slice []any
	mu    sync.Mutex
	cond  *sync.Cond
}

// NewSafeSlice creates and initializes a new instance of a thread-safe slice.
func NewSafeSlice() *SafeSlice {
	ss := &SafeSlice{}
	ss.cond = sync.NewCond(&ss.mu)
	return ss
}

// Add appends a value to the slice in a thread-safe manner.
func (ss *SafeSlice) Add(value any) {
	ss.mu.Lock()
	ss.slice = append(ss.slice, value)
	// After adding a new item, wake up one or all waiting goroutines.
	ss.cond.Broadcast()
	ss.mu.Unlock()
}

// Get retrieves a value by index from the slice in a thread-safe manner.
// Returns the value and true if the index is within bounds; otherwise nil and false.
func (ss *SafeSlice) Get(index int) (any, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if index < 0 || index >= len(ss.slice) {
		return nil, false
	}
	return ss.slice[index], true
}

// GetWait retrieves a value by index from the slice, blocking until the value is available.
func (ss *SafeSlice) GetWait(index int) any {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	// Wait until the index is within the bounds of the slice.
	for len(ss.slice) <= index {
		ss.cond.Wait()
	}
	return ss.slice[index]
}

// Flush clears the slice, resetting it to an empty state.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	ss.slice = []any{}
	ss.mu.Unlock()
}
