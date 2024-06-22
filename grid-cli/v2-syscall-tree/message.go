package v2

// Message structure with promise as the first element of Parms
type Message struct {
	Parms   []interface{}          `json:"parms"` // First element is the promise
	Payload map[string]interface{} `json:"payload"`
}
