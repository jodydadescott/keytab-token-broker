/*
Copyright © 2020 Jody Scott <jody@thescottsweb.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package keytabs

import (
	"encoding/json"

	"github.com/jinzhu/copier"
)

// Keytab contain credentials in the form of a username (or principal) and an
// encrypted password. Keytabs are used to prove identity specifically for
// services and scripts.
type Keytab struct {
	Principal  string `json:"principal,omitempty" yaml:"principal,omitempty"`
	Base64File string `json:"base64file,omitempty" yaml:"base64file,omitempty"`
	SoftExp    int64  `json:"softExp,omitempty" yaml:"softExp,omitempty"`
	HardExp    int64  `json:"hardExp,omitempty" yaml:"hardExp,omitempty"`
}

// JSON Return JSON String representation
func (t *Keytab) JSON() string {
	j, _ := json.Marshal(t)
	return string(j)
}

// Clone return copy of entity
func (t *Keytab) Clone() *Keytab {
	clone := &Keytab{}
	copier.Copy(&clone, &t)
	return clone
}
