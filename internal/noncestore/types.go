package noncestore

import (
	"encoding/json"
	"errors"
	"time"
)

// ErrNotFound ...
var ErrNotFound error = errors.New("Not Found")

// Nonce ...
type Nonce struct {
	Exp   int64  `json:"exp,omitempty"`
	Value string `json:"value,omitempty"`
}

// NewNonce ...
func NewNonce() *Nonce {
	return &Nonce{}
}

// JSON ...
func (t *Nonce) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// NonceFromJSON ...
func NonceFromJSON(b []byte) (*Nonce, error) {
	var t Nonce
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Valid ...
func (t *Nonce) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}
