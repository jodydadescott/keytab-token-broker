package keytabstore

import (
	"encoding/json"
	"time"
)

// Keytab Kerberos Keytab
type Keytab struct {
	Principal  string `json:"principal,omitempty" yaml:"principal,omitempty"`
	Base64File string `json:"base64file,omitempty" yaml:"base64file,omitempty"`
	Exp        int64  `json:"exp,omitempty" yaml:"exp,omitempty"`
}

// JSON return JSON string representation of entity
func (t *Keytab) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// Valid Returns true if entity is valid
func (t *Keytab) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}
