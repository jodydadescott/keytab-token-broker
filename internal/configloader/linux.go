// +build linux

package configloader

import (
	"bufio"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
)

var staticRuntimeConfigFile = "/.kbridge-runtime"

// GetRuntimeConfigString ...
func GetRuntimeConfigString() (string, error) {

	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	f, err := os.Open(home + staticRuntimeConfigFile)
	if err != nil {
		return "", err
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// SetRuntimeConfigString ...
func SetRuntimeConfigString(runtimeConfigString string) error {

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(home+staticRuntimeConfigFile, []byte(runtimeConfigString), 0644)
}
