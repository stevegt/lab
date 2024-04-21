package multistage

import (
	"testing"
)

// testSafeSlice is designed to fully exercise the safeSlice's functionality in a multithreaded scenario
func TestSafeSlice(t *testing.T) {
	ss := &safeSlice{}

	// Add elements to the safeSlice in a separate goroutine to simulate concurrent access
	go func() {
		for i := 0; i < 10; i++ {
			ss.Add(i)
		}
		ss.Add("done") // Signify completion of adds
	}()

	// Attempt to retrieve elements
	var retrieved []any
	for i := 0; i <= 10; i++ {
		value, ok := ss.Get(int64(i))
		if !ok {
			t.Errorf("Failed to get value at index %d", i)
		}
		retrieved = append(retrieved, value)
	}

	// Check for the "done" signal to ensure all adds completed
	if retrieved[len(retrieved)-1] != "done" {
		t.Errorf("Did not retrieve all expected items, last item retrieved: %v", retrieved[len(retrieved)-1])
	}

	// Ensure GetWait works as expected by adding a value after a slight delay
	go func() {
		ss.Add("async")
	}()
	value := ss.GetWait(11) // Waiting for the 11th value
	if value != "async" {
		t.Errorf("GetWait did not return the expected 'async' value, got: %v", value)
	}

	// Test the Flush functionality
	ss.Flush()
	if len(ss.slice) != 0 {
		t.Errorf("Flush did not empty the slice, remaining items: %v", ss.slice)
	}

}
