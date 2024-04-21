package multistage

import (
	"sync"
)

// SafeSlice provides a concurrency-safe implementation for managing a dynamic slice.
// It ensures that all operations on the slice are safe to use concurrently through internal synchronization mechanisms.
type SafeSlice struct {
	mu       sync.RWMutex         // Protects access to the internal slice structure.
	slice    []any                // The underlying slice that stores the actual data.
	wChan    chan Element         // A channel for concurrent writes to the SafeSlice.
	getChans map[int]chan<- any   // A map tracking channels waiting for data by index.
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
		wChan:    make(chan Element, 1024), // Buffered channel to reduce blocking on writes.
		getChans: make(map[int]chan<- any),
	}
	go ss.listen() // Start the goroutine that listens for write operations and serves data requests.
	return ss
}

// listen is a goroutine that handles write operations to the slice and serves data requests.
// It ensures that all operations are performed in a safe, synchronized manner.
func (ss *SafeSlice) listen() {
	for el := range ss.wChan {
		ss.mu.Lock()
		if el.Index >= 0 && el.Index < len(ss.slice) {
			ss.slice[el.Index] = el.Value // Update existing index.
		} else if el.Index == len(ss.slice) {
			ss.slice = append(ss.slice, el.Value) // Append to the end.
		}
		if ch, ok := ss.getChans[el.Index]; ok {
			ch <- el.Value
			delete(ss.getChans, el.Index)
		}
		ss.mu.Unlock()
	}
}

// Append adds a new element to the end of the SafeSlice, using its concurrent write mechanism.
func (ss *SafeSlice) Append(value any) {
	ss.mu.RLock()
	index := len(ss.slice)
	ss.mu.RUnlock()
	ss.wChan <- Element{Index: index, Value: value}
}

// Replace updates the element at the given index within the SafeSlice, if the index exists.
func (ss *SafeSlice) Replace(index int, value any) {
	ss.wChan <- Element{Index: index, Value: value}
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
}

// GetChan provides a way to asynchronously receive an element from the SafeSlice by index.
// It returns a read-only channel that will either receive the requested element or remain empty.
func (ss *SafeSlice) GetChan(index int) <-chan any {
	ch := make(chan any, 1) // Buffered channel for non-blocking send.
	ss.mu.Lock()
	if index >= 0 && index < len(ss.slice) {
		ch <- ss.slice[index]
	} else {
		ss.getChans[index] = ch
	}
	ss.mu.Unlock()
	return ch
}
