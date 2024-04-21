package multistage

import (
	"sync"
)

// SafeSlice provides a concurrency-safe implementation for a dynamic array.
type SafeSlice struct {
	mu     sync.Mutex              // Protects access to the internal slice
	slice  []any                   // The internal dynamic array
	wChan  chan Element            // Channel to write new elements to the SafeSlice
}

// Element represents an element that can be added to the SafeSlice.
type Element struct {
	Value any // The value of the element, with no Index field as per test requirements
}

// NewSafeSlice initializes a new SafeSlice with an internal write channel.
func NewSafeSlice() *SafeSlice {
	ss := &SafeSlice{
		slice: make([]any, 0),
		wChan: make(chan Element, 1024), // Buffered channel for concurrent writes
	}
	go ss.listenForWrite() // Start listening on the write channel
	return ss
}

// listenForWrite listens on the write channel and adds elements to the slice.
func (ss *SafeSlice) listenForWrite() {
	for elem := range ss.wChan {
		ss.mu.Lock()
		ss.slice = append(ss.slice, elem.Value)
		ss.mu.Unlock()
	}
}

// Add adds a new element to the SafeSlice by sending it to the write channel.
func (ss *SafeSlice) Add(value any) {
	ss.wChan <- Element{Value: value}
}

// Get safely retrieves an element from the SafeSlice by its index.
func (ss *SafeSlice) Get(index int) (any, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if index < 0 || index >= len(ss.slice) {
		return nil, false
	}
	return ss.slice[index], true
}

// WriteChan provides a channel for writing elements to the SafeSlice.
func (ss *SafeSlice) WriteChan() chan<- Element {
	return ss.wChan
}

// Flush clears the contents of the SafeSlice.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	ss.slice = nil // Clear the slice
	ss.mu.Unlock()
}
