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

package nonces

import (
	"testing"
	"time"
)

func Test1(t *testing.T) {

	var err error

	config := &Config{
		CacheRefreshInterval: 5,
		Lifetime:             5,
	}

	nonces, err := config.Build()

	if err != nil {
		t.Fatalf("Unexpected err %s", err)
	}

	nonce1 := nonces.NewNonce()
	nonce2 := nonces.NewNonce()
	nonce3 := nonces.NewNonce()

	if nonce1.Value == nonce2.Value {
		t.Fatalf("Unexpected")
	}

	if nonce1.Value == nonce3.Value {
		t.Fatalf("Unexpected")
	}

	if nonce2.Value == nonce3.Value {
		t.Fatalf("Unexpected")
	}

	if x := nonces.GetNonce(nonce1.Value); x == nil {
		t.Fatalf("Unexpected")
	}

	if x := nonces.GetNonce(nonce2.Value); x == nil {
		t.Fatalf("Unexpected")
	}

	if x := nonces.GetNonce(nonce3.Value); x == nil {
		t.Fatalf("Unexpected")
	}

	time.Sleep(2 * time.Second)

	if x := nonces.GetNonce(nonce1.Value); x == nil {
		t.Fatalf("Unexpected")
	}

	if x := nonces.GetNonce(nonce2.Value); x == nil {
		t.Fatalf("Unexpected")
	}

	if x := nonces.GetNonce(nonce3.Value); x == nil {
		t.Fatalf("Unexpected")
	}

	time.Sleep(6 * time.Second)

	if x := nonces.GetNonce(nonce1.Value); x != nil {
		t.Fatalf("Unexpected")
	}

	if x := nonces.GetNonce(nonce2.Value); x != nil {
		t.Fatalf("Unexpected")
	}

	if x := nonces.GetNonce(nonce3.Value); x != nil {
		t.Fatalf("Unexpected")
	}

	// var principals []string
	// principals = append(principals, "bob@example.com")
	// principals = append(principals, "alice@example.com")

	// config := &Config{
	// 	Principals: principals,
	// }

	// store, err := config.Build()
	// if err != nil {
	// 	t.Fatalf("Unexpected err %s", err)
	// }
	// defer store.Shutdown()

	// k := store.GetKeytab("bob@example.com")

	// if k == nil {
	// 	t.Fatalf("Unexpected")
	// }

	// k = store.GetKeytab("invalid@example.com")

	// if k != nil {
	// 	t.Fatalf("Unexpected")
	// }

}
