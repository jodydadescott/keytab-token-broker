package model

import (
	"encoding/json"
	"time"
)

// Nonce holds one time expiring secret
type Nonce struct {
	Exp   int64  `json:"exp,omitempty" yaml:"exp,omitempty"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// NewNonce Returns new Entity
func NewNonce() *Nonce {
	return &Nonce{}
}

// JSON Returns JSON string representation of entity
func (t *Nonce) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// NonceFromJSON Returns entity from JSON string
func NonceFromJSON(b []byte) (*Nonce, error) {
	var t Nonce
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Valid Returns true if entity is valid
func (t *Nonce) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}
