// +build linux darwin

package keytabstore

import (
	"fmt"
	"runtime"
)

func osSpecificNewKeytab(principal string) (string, error) {
	return fmt.Sprintf("This is not a valid keytab file because the OS %s is not supported. Only Windows is supported at this time", runtime.GOOS), nil
}
