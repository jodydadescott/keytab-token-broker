package nonces

import (
	"time"
)

// Nonce holds one time expiring secret
type Nonce struct {
	Exp   int64  `json:"exp,omitempty" yaml:"exp,omitempty"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// Valid Returns true if entity is valid
func (t *Nonce) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}
