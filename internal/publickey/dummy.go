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

package publickey

import (
	"fmt"
	"sync"
)

// DummyCache ...
type DummyCache struct {
	mutex    sync.Mutex
	internal map[string]*PublicKey
}

// Dummy Returns a new Dummy PublicKey Cache
func Dummy() Cache {
	return &DummyCache{
		internal: make(map[string]*PublicKey),
	}
}

// PutKey Puts key
func (t *DummyCache) PutKey(key *PublicKey) error {

	if key.Iss == "" {
		return fmt.Errorf("Missing Issuer (iss)")
	}

	if key.Kid == "" {
		return fmt.Errorf("Missing Kid (kid)")
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.internal[key.Iss+":"+key.Kid] = key
	return nil
}

// GetKey Returns PublicKey from cache if found. If not gets PublicKey from
// validated issuer, stores in cache and returns copy
func (t *DummyCache) GetKey(iss, kid string) (*PublicKey, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	key, exist := t.internal[iss+":"+kid]
	if exist {
		return key.Copy(), nil
	}

	return nil, ErrNotFound
}

// Shutdown Cache
func (t *DummyCache) Shutdown() {
}
