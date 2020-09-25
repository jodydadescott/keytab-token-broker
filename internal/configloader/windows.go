// +build windows

package configloader

import (
	"golang.org/x/sys/windows/registry"
)

var keyRegistryPath = `SOFTWARE\KerberosBridge`

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
