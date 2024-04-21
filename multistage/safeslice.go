package multistage

import (
	"sync"
)

// Element struct is now adjusted to conform to the test cases, including an Index field.
// However, the usage of Index will be ignored in current insertion logic as it merely appends items.
type Element struct {
	Index int // The index where the element is intended to be added; usage is context-dependent.
	Value any // The actual value of the element to be added.
}

// SafeSlice struct definition remains unchanged, focusing on providing a thread-safe slice management mechanism.
type SafeSlice struct {
	mu       sync.Mutex
	slice    []any
	getChans map[int][]chan any
	wChan    chan Element
}

// NewSafeSlice initializes a SafeSlice with default values and starts
// a listening goroutine for wChan.
func NewSafeSlice() *SafeSlice {
	ss := &SafeSlice{
		getChans: make(map[int][]chan any),
		wChan:    make(chan Element, 10), // Use a buffered channel for non-blocking sends.
	}
	go ss.processAdds()
	return ss
}

// WriteChan returns a channel that can be used to write elements to the SafeSlice.
func (ss *SafeSlice) WriteChan() chan<- Element {
	return ss.wChan
}

// Add immediately sends an element to the wChan for concurrent-safe addition.
func (ss *SafeSlice) Add(value any) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	index := len(ss.slice)
	ss.wChan <- Element{Index: index, Value: value}
}

// Get allows for retrieving an element by index, checking boundaries safely.
func (ss *SafeSlice) Get(index int) (value any, ok bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if index >= 0 && index < len(ss.slice) {
		return ss.slice[index], true
	}
	return nil, false
}

// Flush clears the slice and its dependencies for reuse or disposal.
func (ss *SafeSlice) Flush() {
	ss.mu.Lock()
	ss.slice = []any{}
	ss.getChans = make(map[int][]chan any)
	ss.mu.Unlock()
}

// GetChan waits for or immediately retrieves an element at a given index, depending on availability.
func (ss *SafeSlice) GetChan(index int) chan any {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ch := make(chan any, 1)
	if index < len(ss.slice) {
		ch <- ss.slice[index]
		close(ch)
	} else {
		ss.getChans[index] = append(ss.getChans[index], ch)
	}
	return ch
}

// processAdds listens on wChan for new elements, appending them to the slice and handling waiters.
func (ss *SafeSlice) processAdds() {
	for elem := range ss.wChan {
		ss.mu.Lock()
		ss.slice = append(ss.slice, elem.Value)
		if waiting, exists := ss.getChans[len(ss.slice)-1]; exists {
			for _, ch := range waiting {
				ch <- elem.Value
				close(ch)
			}
			delete(ss.getChans, len(ss.slice)-1)
		}
		ss.mu.Unlock()
	}
}
