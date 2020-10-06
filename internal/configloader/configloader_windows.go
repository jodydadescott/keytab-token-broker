// +build windows

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

package configloader

import (
	"golang.org/x/sys/windows/registry"
)

var keyRegistryPath = `SOFTWARE\KTBServer`

// GetRuntimeConfigString ...
func GetRuntimeConfigString() (string, error) {

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyRegistryPath, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	runtimeConfigString, _, err := k.GetStringValue("RuntimeConfigString")
	if err != nil {
		if err != registry.ErrNotExist {
			return "", nil
		}
		return "", err
	}

	return runtimeConfigString, nil
}

// SetRuntimeConfigString ...
func SetRuntimeConfigString(runtimeConfigString string) error {

	// _ arg is if key already existed
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyRegistryPath, registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()

	err = k.SetStringValue("RuntimeConfigString", runtimeConfigString)
	if err != nil {
		return err
	}

	return nil
}
