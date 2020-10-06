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

package token

import (
	"encoding/json"

	"github.com/jinzhu/copier"
)

// Token OAUTH/OIDC Token
type Token struct {
	TokenString string                 `json:"tokenString,omitempty" yaml:"tokenString,omitempty"`
	Alg         string                 `json:"alg,omitempty" yaml:"alg,omitempty"`
	Kid         string                 `json:"kid,omitempty" yaml:"kid,omitempty"`
	Iss         string                 `json:"iss,omitempty" yaml:"iss,omitempty"`
	Exp         int64                  `json:"exp,omitempty" yaml:"exp,omitempty"`
	Aud         string                 `json:"aud,omitempty" yaml:"aud,omitempty"`
	Claims      map[string]interface{} `json:"claims,omitempty" yaml:"claims,omitempty"`
}

// JSON ...
func (t *Token) JSON() string {
	j, _ := json.Marshal(t)
	return string(j)
}

// Clone return copy
func (t *Token) Clone() *Token {
	clone := &Token{}
	copier.Copy(&clone, &t)
	return clone
}
