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

package policy

// Decision ...
type Decision struct {
	Auth       bool
	Principals []string
}

// HasPrincipal Returns true if principal is present in entity
func (t *Decision) HasPrincipal(principal string) bool {
	if principal == "" {
		return false
	}
	for _, s := range t.Principals {
		if s == principal {
			return true
		}
	}
	return false
}
