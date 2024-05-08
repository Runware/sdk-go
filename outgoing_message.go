package runware

import (
	"encoding/json"
)

// OutgoingMessageHandler interface implementation for outgoing messages
type OutgoingMessageHandler interface {
	MarshalBinary() ([]byte, error)
}

// NewConnection message
type NewConnection struct {
	APIKey string `json:"apiKey"`
}

func (msg *NewConnection) MarshalBinary() ([]byte, error) {
	binMsg := map[string]NewConnection{
		"newConnection": {
			APIKey: msg.APIKey,
		},
	}
	
	return json.Marshal(binMsg)
}
