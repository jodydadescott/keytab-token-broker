// +build  darwin

package configloader

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap/zapcore"
)

var staticRuntimeConfigFile = "/Library/kbridge-runtime.conf"

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

func getZapHook() (func(zapcore.Entry) error, error) {
	return nil, fmt.Errorf("Not supported")
}
