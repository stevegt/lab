package multistage

import (
	"sync"
)

// SafeSlice is a thread-safe slice supporting concurrent operations.
type SafeSlice struct {
	slice []any
	mu    sync.Mutex
	waiters map[int][]chan any
}

// NewSafeSlice creates and initializes a new instance of a thread-safe slice.
func NewSafeSlice() *SafeSlice {
	return &SafeSlice{
		waiters: make(map[int][]chan any),
	}
}

// Add appends a value to the slice in a thread-safe manner.
func (ss *SafeSlice) Add(value any) {
	ss.mu.Lock()
	index := len(ss.slice)
	ss.slice = append(ss.slice, value)
	if waiters, ok := ss.waiters[index]; ok {
		for _, waiter := range waiters {
			waiter <- value
			close(waiter)
		}
		delete(ss.waiters, index)
	}
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

// Flush clears the slice, resetting it to an empty state.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	ss.slice = []any{}
	ss.waiters = make(map[int][]chan any)
	ss.mu.Unlock()
}

// GetChan creates or returns an existing channel for the specified index.
// If the index's value is not yet available, it will return a channel that will be resolved when the value is added.
func (ss *SafeSlice) GetChan(index int) chan any {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if index >= 0 && index < len(ss.slice) {
		ch := make(chan any, 1)
		ch <- ss.slice[index]
		close(ch)
		return ch
	} else {
		ch := make(chan any, 1)
		ss.waiters[index] = append(ss.waiters[index], ch)
		return ch
	}
}
