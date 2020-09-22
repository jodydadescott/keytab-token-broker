// +build windows

package cmd

import (
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/eventlog"
)

var keyRegistryPath = `SOFTWARE\KerberosBridge`

func getRuntimeConfigString() (string, error) {

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

func setRuntimeConfigString(runtimeConfigString string) error {

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

var wlog *eventlog.Log

func getZapHook() (func(zapcore.Entry) error, error) {

	if wlog == nil {

		//Set the log source name which will appear in the Windows Event Log
		var loggerName = "KerberosBridge"

		//Setup Windows Event log with the log source name and logging levels
		err := eventlog.InstallAsEventCreate(loggerName, eventlog.Info|eventlog.Warning|eventlog.Error)

		if err != nil {
			return nil, err
		}

		//Open a handler to the event logger
		wlog, err = eventlog.Open(loggerName)
		if err != nil {
			return nil, err
		}

	}

	return func(e zapcore.Entry) error {

		// 	Level      Level
		// 	Time       time.Time
		// 	LoggerName string
		// 	Message    string
		// 	Caller     EntryCaller
		// 	Stack      string

		// Level
		// DebugLevel
		// InfoLevel
		// WarnLevel
		// ErrorLevel
		// DPanicLevel
		// PanicLevel
		// FatalLevel

		switch e.Level {

		case zapcore.InfoLevel:
			wlog.Info(1, e.Message)
			break

		case zapcore.WarnLevel:
			wlog.Warning(1, e.Message)
			break

		// Everthing else
		default:
			wlog.Error(1, e.Message)
			break

		}

		return nil
	}, nil
}
