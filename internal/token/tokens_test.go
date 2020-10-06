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
	"reflect"
	"testing"
)

// TODO Remove these tokens and have the test pull current tokens

// func Test1(t *testing.T) {

// 	config := NewConfig()

// 	s, err := NewTokenStore(config)
// 	if err != nil {
// 		t.Fatalf("Error not expected: %s", err)
// 	}
// 	defer s.Shutdown()

// 	validExpiredToken, err := s.GetToken(validExpiredTokenStr)

// 	if err != nil {
// 		t.Fatalf("Error not expected: %s", err)
// 	}

// 	assertEqual(t, validExpiredToken.Iss, "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo")

// 	_, err = s.GetToken(invalidTokenStr)

// 	if err == nil {
// 		t.Fatalf("Expected error for invalid token")
// 	}

// }

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	// debug.PrintStack()
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}
