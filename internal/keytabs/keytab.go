package keytabs

import (
	"time"
)

// Keytab Kerberos Keytab
type Keytab struct {
	Principal  string `json:"principal,omitempty" yaml:"principal,omitempty"`
	Base64File string `json:"base64file,omitempty" yaml:"base64file,omitempty"`
	Exp        int64  `json:"exp,omitempty" yaml:"exp,omitempty"`
}

// Valid Returns true if entity is valid
func (t *Keytab) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}
