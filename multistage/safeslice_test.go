package multistage

import (
	"fmt"
	"testing"
)

func testSafeSlice(t *testing.T) error {
	// Create a new SafeSlice
	ss := &safeSlice{}

	// Add a new element to the SafeSlice
	ss.Add("Hello")

	// Get the first element from the SafeSlice
	_, _ = ss.Get(0)

	return fmt.Errorf("needs more testing")
}
