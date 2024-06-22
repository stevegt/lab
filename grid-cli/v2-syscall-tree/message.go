package v2

// Message defines the structure for communication messages
type Message struct {
	Parms    []interface{}          `json:"parms"`
	Payload  map[string]interface{} `json:"payload"`
}
