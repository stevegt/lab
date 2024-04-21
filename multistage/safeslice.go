package multistage

import (
	"sync"
)

// safeSlice is a thread-safe slice implementation using a mutex for synchronization.
type safeSlice struct {
	slice []any       // slice stores the data in a slice form
	mu    sync.Mutex  // mu protects access to the slice
	wg    sync.WaitGroup // wg is used for signaling the availability of new elements
}

// Add appends a value to the safeSlice, ensuring thread-safety via mutex lock.
func (s *safeSlice) Add(value any) {
	s.mu.Lock()
	s.slice = append(s.slice, value)
	s.mu.Unlock()
	s.wg.Done() // Signal that a new element has been added
}

// Get retrieves a value at a given index from the safeSlice if the index is valid.
func (s *safeSlice) Get(index int64) (value any, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if index is out of bounds
	if index < 0 || int(index) >= len(s.slice) {
		return nil, false
	}

	return s.slice[index], true
}

// GetWait retrieves a value at a given index from the safeSlice, waiting if necessary.
func (s *safeSlice) GetWait(index int64) any {
	var value any
	for {
		s.mu.Lock()
		if int(index) < len(s.slice) {
			value = s.slice[index]
			s.mu.Unlock()
			break
		}
		s.mu.Unlock()
		s.wg.Add(1) // Prepare to wait for the element to be added
		s.wg.Wait() // Wait for the signal that a new element is added
	}
	return value
}

// Flush clears all elements from the safeSlice.
func (s *safeSlice) Flush() {
	s.mu.Lock()
	s.slice = []any{} // Reset the slice to an empty slice
	s.mu.Unlock()
}
