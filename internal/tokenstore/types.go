package tokenstore

import (
	"encoding/json"
	"errors"
	"time"
)

// ErrExpired ...
var ErrExpired error = errors.New("Expired")

// ErrNotFound ...
var ErrNotFound error = errors.New("Not Found")

// Token OAUTH/OIDC Token
type Token struct {
	Alg     string   `json:"alg,omitempty"`
	Kid     string   `json:"kid,omitempty"`
	Iss     string   `json:"iss,omitempty"`
	Exp     int64    `json:"exp,omitempty"`
	Aud     string   `json:"aud,omitempty"`
	Keytabs []string `json:"keytabs,omitempty"`
}

// NewToken Returs new Token
func NewToken() *Token {
	return &Token{
		Keytabs: []string{},
	}
}

// JSON returns JSON
func (t *Token) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// TokenFromJSON ...
func TokenFromJSON(b []byte) (*Token, error) {
	var t Token
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Valid returns true if valid
func (t *Token) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}
