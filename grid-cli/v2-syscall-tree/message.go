package v2

// Message defines the structure for communication messages.
// The first element in Parms is the promise of handling or not handling the syscall.
type Message struct {
	Parms    []interface{}          `json:"parms"`   // Parameters: first element is the promise
	Payload  map[string]interface{} `json:"payload"` // Metadata about the promise
}
