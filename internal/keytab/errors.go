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

package keytab

import "errors"

var (
	// ErrNotFound Secret not found
	ErrNotFound error = errors.New("Keytab does not exist")

	// ErrNotReady Keytab exist but has not been processed yet
	ErrNotReady error = errors.New("Keytab exist but has not been processed")

	// ErrGenFail Error occured while attempting to generate Keytab
	ErrGenFail error = errors.New("Error occured while attempting to generate Keytab")
)
