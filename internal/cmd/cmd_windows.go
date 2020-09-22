// +build windows

package cmd

import (
	"fmt"

	"github.com/jodydadescott/kerberos-bridge/internal/server"
	"golang.org/x/sys/windows/registry"
)

var keyRegistryPath = `SOFTWARE\KerberosBridge`

func getConfigFromRegistry() (*server.Config, error) {

	config := server.NewConfig()

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyRegistryPath, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer k.Close()

	// 1
	if x, _, err := k.GetStringValue("LogLevel"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.LogLevel = x
	}

	// 2
	if x, _, err := k.GetStringValue("LogFormat"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.LogFormat = x
	}

	// 3
	if x, _, err := k.GetStringsValue("LogTo"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.LogTo = x
	}

	// 4
	if x, _, err := k.GetStringValue("Listen"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.Listen = x
	}

	// 5
	if x, _, err := k.GetIntegerValue("HTTPPort"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.HTTPPort = int(x)
	}

	// 6
	if x, _, err := k.GetIntegerValue("HTTPSPort"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.HTTPSPort = int(x)
	}

	// 7
	if x, _, err := k.GetIntegerValue("NonceLifetime"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.Nonce.Lifetime = int(x)
	}

	// 8
	if x, _, err := k.GetStringValue("PolicyQuery"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.Policy.Query = x
	}

	// 9
	if x, _, err := k.GetStringValue("PolicyRegoScript"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.Policy.RegoScript = x
	}

	// 10
	if x, _, err := k.GetIntegerValue("KeytabLifetime"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.Keytab.Lifetime = int(x)
	}

	// 11
	if x, _, err := k.GetStringsValue("KeytabPrincipals"); err != nil {
		if err != registry.ErrNotExist {
			return nil, err
		}
	} else {
		config.Keytab.Principals = x
	}

	return config, nil
}

func setRegistryConfig(config *server.Config) error {

	// _ arg is if key already existed
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyRegistryPath, registry.WRITE)
	if err != nil {
		fmt.Println("trace2")
		return err
	}
	defer k.Close()

	// 1
	if config.LogLevel != "" {
		err = k.SetStringValue("LogLevel", config.LogLevel)
		if err != nil {
			return err
		}
	}

	// 2
	if config.LogFormat != "" {
		err = k.SetStringValue("LogFormat", config.LogFormat)
		if err != nil {
			return err
		}
	}

	// 3
	if config.LogTo != nil && len(config.LogTo) > 0 {
		err = k.SetStringsValue("LogTo", config.LogTo)
		if err != nil {
			return err
		}
	}

	// 4
	if config.Listen != "" {
		err = k.SetStringValue("Listen", config.Listen)
		if err != nil {
			return err
		}
	}

	// 5
	if config.HTTPPort > 0 {
		err = k.SetDWordValue("HTTPPort", uint32(config.HTTPPort))
		if err != nil {
			return err
		}
	}

	// 6
	if config.HTTPSPort > 0 {
		err = k.SetDWordValue("HTTPSPort", uint32(config.HTTPSPort))
		if err != nil {
			return err
		}
	}

	// 7
	if config.Nonce.Lifetime > 0 {
		err = k.SetDWordValue("NonceLifetime", uint32(config.Nonce.Lifetime))
		if err != nil {
			return err
		}
	}

	// 8
	if config.Policy.Query != "" {
		err = k.SetStringValue("PolicyQuery", config.Policy.Query)
		if err != nil {
			return err
		}
	}

	// 9
	if config.Policy.RegoScript != "" {
		err = k.SetStringValue("PolicyRegoScript", config.Policy.RegoScript)
		if err != nil {
			return err
		}
	}

	// 10
	if config.Keytab.Lifetime > 0 {
		err = k.SetDWordValue("KeytabLifetime", uint32(config.Keytab.Lifetime))
		if err != nil {
			return err
		}
	}

	// 11
	if config.Keytab.Principals != nil && len(config.Keytab.Principals) > 0 {
		err = k.SetStringsValue("KeytabPrincipals", config.Keytab.Principals)
		if err != nil {
			return err
		}
	}

	return nil

}

func getRuntimeConfigString() (string, error) {

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyRegistryPath, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	bootConfig, _, err := k.GetStringValue("BootConfig")
	if err != nil {
		if err != registry.ErrNotExist {
			return "", nil
		}
		return "", err
	}

	return bootConfig, nil
}

func setRuntimeConfigString(bootConfig string) error {

	// _ arg is if key already existed
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, keyRegistryPath, registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()

	err = k.SetStringValue("BootConfig", bootConfig)
	if err != nil {
		return err
	}

	return nil
}
