package multistage

import (
	"sync"
)

// SafeSlice is a thread-safe slice supporting concurrent operations.
type SafeSlice struct {
	slice []any
	mu    sync.Mutex
	// cond  *sync.Cond
	waitchan chan chan bool
}

// NewSafeSlice creates and initializes a new instance of a thread-safe slice.
func NewSafeSlice() *SafeSlice {
	ss := &SafeSlice{}
	// ss.cond = sync.NewCond(&ss.mu)
	ss.waitchan = make(chan chan bool)
	return ss
}

// Add appends a value to the slice in a thread-safe manner.
func (ss *SafeSlice) Add(value any) {
	ss.mu.Lock()
	ss.slice = append(ss.slice, value)
	// After adding a new item, wake up one or all waiting goroutines.
	for c := range ss.waitchan {
		c <- true
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

/*
// GetWait retrieves a value by index from the slice, blocking until
// the value is available or a timeout occurs.
func (ss *SafeSlice) GetWait(index int, timeout time.Duration) (any, bool) {
	// Wait until the index is within the bounds of the slice.
	for len(ss.slice) <= index {
		// notify Add that we are waiting
		c := make(chan bool)
		ss.waitchan <- c
		// use select with timeout
		select {
		case <-time.After(timeout):
			return nil, false
		case <-c:
			co
		}
		/ get lock
		if ss.mu.TryLock() {



	// get lock
	for len(ss.slice) <= index {
	}

	return ss.slice[index]
		XXX ss.cond.Wait()
	XXX TryLock or understand Cond.Wait or just use channel
}
*/

// Flush clears the slice, resetting it to an empty state.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	ss.slice = []any{}
	ss.mu.Unlock()
}
