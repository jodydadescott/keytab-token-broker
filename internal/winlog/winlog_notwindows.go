// +build windows

package winlog

import (
	"golang.org/x/sys/windows/svc/eventlog"
)

var wlog *eventlog.Log
