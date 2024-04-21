package multistage

import "sync"

// safeSlice is a thread-safe slice that uses a mutex to synchronize
// access.
type safeSlice struct {
	slice []any
	mu    sync.Mutex
	wg    sync.WaitGroup
}

// Add appends a value to the slice, locking the mutex to ensure thread safety
func (s *safeSlice) Add(value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slice = append(s.slice, value)
}

// Get safely retrieves a value from the slice by index
func (s *safeSlice) Get(index int64) (any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= int64(len(s.slice)) {
		return 0, false
	}
	return s.slice[index], true
}

// GetWait safely retrieves a value from the slice by index, waiting
// until the index is valid.
func (s *safeSlice) GetWait(index int64) any {
	for {
		s.mu.Lock()
		value, ok := s.Get(index)
		if ok {
			s.mu.Unlock()
			return value
		}
		s.wg.Add(1)
		s.mu.Unlock()
		s.wg.Wait()
	}
}

// Flush resets the slice to an empty state
func (s *safeSlice) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slice = s.slice[:0]
}
