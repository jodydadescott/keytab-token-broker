// +build linux darwin

package cmd

import (
	"fmt"

	"github.com/jodydadescott/kerberos-bridge/internal/server"
)

func getConfigFromRegistry() (*server.Config, error) {
	return nil, fmt.Errorf("Implemented on Windows only")
}

func setRegistryConfig(*server.Config) error {
	return fmt.Errorf("Implemented on Windows only")
}

func getRuntimeConfigString() (string, error) {
	return "", fmt.Errorf("Implemented on Windows only")
}

func setRuntimeConfigString(string) error {
	return fmt.Errorf("Implemented on Windows only")
}
