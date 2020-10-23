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

import "errors"

var (
	// ErrInvalid Invalid format
	ErrInvalid error = errors.New("Invalid format")

	// ErrMissingField Token is missing a required field
	ErrMissingField error = errors.New("Token is missing a required field")

	// ErrSignatureInvalid Token signature is invalid
	ErrSignatureInvalid error = errors.New(" Token signature is invalid")

	// ErrExpired Token is expired
	ErrExpired error = errors.New("Token is expired")

	// ErrNonceInvalid Token Nonce is not valid
	ErrNonceInvalid error = errors.New("Token Nonce is invalid")
)
