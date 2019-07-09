package transaction

import (
	"encoding/json"

	"github.com/aergoio/aergo-esindexer/types"
)

// Payload is an unmarshalled contract call payload
type Payload struct {
	Name string   `json:"Name"`
	Args []string `json:"Args"`
}

// UnmarshalPayload concerts payload bytes into a struct using json
func UnmarshalPayload(tx *types.Tx) (*Payload, error) {
	payloadSource := tx.GetBody().GetPayload()
	payload := new(Payload)
	err := json.Unmarshal(payloadSource, payload)
	if err != nil {
		return &Payload{}, err
	}
	return payload, nil
}

// GetCallName extracts the contract call function name from the payload
func GetCallName(tx *types.Tx) (string, error) {
	payload, err := UnmarshalPayload(tx)
	if err != nil {
		return "", err
	}
	return payload.Name, nil
}
