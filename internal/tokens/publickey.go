package tokens

import (
	"crypto/ecdsa"
	"encoding/json"

	"github.com/jinzhu/copier"
)

// PublicKey ...
type PublicKey struct {
	EcdsaPublicKey *ecdsa.PublicKey `json:"-"`
	Iss            string           `json:"iss,omitempty" yaml:"iss,omitempty"`
	Kid            string           `json:"kid,omitempty" yaml:"kid,omitempty"`
	Kty            string           `json:"kty,omitempty" yaml:"kty,omitempty"`
	Exp            int64            `json:"exp,omitempty" yaml:"exp,omitempty"`
}

// Expiration ...
func (t *PublicKey) Expiration() int64 {
	return t.Exp
}

// HardExpiration ...
func (t *PublicKey) HardExpiration() int64 {
	return t.Exp
}

// Key ...
func (t *PublicKey) Key() string {
	return t.Iss + t.Kid
}

// JSON ...
func (t *PublicKey) JSON() string {
	j, _ := json.Marshal(t)
	return string(j)
}

// Clone return copy
func (t *PublicKey) Clone() *PublicKey {
	clone := &PublicKey{}
	copier.Copy(&clone, &t)
	return clone
}
