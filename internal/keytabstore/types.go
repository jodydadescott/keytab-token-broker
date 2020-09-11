package keytabstore

import (
	"encoding/json"
	"regexp"
	"time"
)

var (
	principalRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Keytab Kerberos Keytab
type Keytab struct {
	Principal  string `json:"principal,omitempty"`
	Base64File string `json:"base64file,omitempty"`
	Exp        int64  `json:"exp,omitempty"`
	//UnSealed     bool   `json:"-"`
}

// JSON return JSON string representation of instance
func (t *Keytab) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// KeytabFromJSON ...
func KeytabFromJSON(b []byte) (*Keytab, error) {
	var t Keytab
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// AssertTypeValid Assert type is valid
func (t *Keytab) AssertTypeValid() error {
	return nil
}

// Valid returns true if valid
func (t *Keytab) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}

// // Principal ...
// type Principal struct {
// 	Principal string `json:"principal,omitempty"`
// }

// // JSON return JSON string representation of instance
// func (t *Principal) JSON() string {
// 	e, err := json.Marshal(t)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	return string(e)
// }

// // PrincipalFromJSON Return new instance from JSON string
// func PrincipalFromJSON(b []byte) (*Principal, error) {
// 	var t Principal
// 	err := json.Unmarshal(b, &t)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &t, nil
// }

// // AssertTypeValid Assert type is valid
// func (t *Principal) AssertTypeValid() error {

// 	e := []string{}

// 	if t.Principal == "" {
// 		e = append(e, "Principal is empty")
// 	}

// 	if len(t.Principal) < 3 && len(t.Principal) > 254 {
// 		if len(t.Principal) < 3 {
// 			e = append(e, "Principal is to short")
// 		} else {
// 			e = append(e, "Principal is to long")
// 		}
// 	}

// 	if !principalRegex.MatchString(t.Principal) {
// 		e = append(e, "Principal is invalid")
// 	}

// 	if len(e) > 0 {
// 		return fmt.Errorf("Invalid: %s", e)
// 	}
// 	return nil
// }
