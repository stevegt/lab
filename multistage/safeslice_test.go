package multistage

import (
	"testing"
	"time"
)

// testSafeSlice is designed to fully exercise the SafeSlice's functionality in a multithreaded scenario.
func TestSafeSlice(t *testing.T) {
	ss := NewSafeSlice()

	// Initialize a waiting mechanism to ensure all adds are completed before validations.
	doneAdding := make(chan bool)

	// Add elements to the safeSlice in a separate goroutine to simulate concurrent access.
	go func() {
		for i := 0; i < 10; i++ {
			ss.Add(i)
		}
		ss.Add("done") // Signify completion of adds.
		doneAdding <- true
	}()

	<-doneAdding // Wait for signal that adds are done.

	// Attempt to retrieve elements.
	var retrieved []any
	// Use GetWait to ensure synchronization with element addition.
	for i := 0; i < 11; i++ {
		value := ss.GetWait(i)
		retrieved = append(retrieved, value)
	}

	// Check for the "done" signal to ensure all adds completed.
	if retrieved[len(retrieved)-1] != "done" {
		t.Errorf("Did not retrieve all expected items, last item retrieved: %v", retrieved[len(retrieved)-1])
	}

	// Ensure GetWait works as expected by adding a value after a slight delay.
	go func() {
		time.Sleep(100 * time.Millisecond) // Ensure this operation happens after the previous retrieval.
		ss.Add("async")
	}()
	value := ss.GetWait(11) // Waiting for the 11th value.
	if value != "async" {
		t.Errorf("GetWait did not return the expected 'async' value, got: %v", value)
	}

	// Test the Flush functionality.
	ss.Flush()
	if len(ss.slice) != 0 {
		t.Errorf("Flush did not empty the slice, remaining items: %v", ss.slice)
	}
}
