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

// +build windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jodydadescott/keytab-token-broker/internal/configloader"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const svcName = "KTBServer"
const svcDesc = "Keytab Token Broker Server"

var elog debug.Log

func installService() error {
	exepath, err := exePath()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(svcName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", svcName)
	}
	// Jody
	// s, err = m.CreateService(svcName, exepath, mgr.Config{DisplayName: svcDesc}, "is", "auto-started")
	s, err = m.CreateService(svcName, exepath, mgr.Config{DisplayName: svcDesc})
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(svcName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func removeService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(svcName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", svcName)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(svcName)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func stopService() error {
	return controlService(svc.Stop, svc.Stopped)
}

func pauseService() error {
	return controlService(svc.Pause, svc.Paused)
}

func continueService() error {
	return controlService(svc.Continue, svc.Running)
}

func startService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(svcName)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func isAnInteractiveSession() (bool, error) {
	return svc.IsAnInteractiveSession()
}

func runService() {
	var err error

	elog, err = eventlog.Open(svcName)
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", svcName))
	run := svc.Run

	// if isDebug {
	// 	run = debug.Run
	// }

	err = run(svcName, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", svcName, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", svcName))
}

// const (
// 	Stopped         = State(windows.SERVICE_STOPPED)
// 	StartPending    = State(windows.SERVICE_START_PENDING)
// 	StopPending     = State(windows.SERVICE_STOP_PENDING)
// 	Running         = State(windows.SERVICE_RUNNING)
// 	ContinuePending = State(windows.SERVICE_CONTINUE_PENDING)
// 	PausePending    = State(windows.SERVICE_PAUSE_PENDING)
// 	Paused          = State(windows.SERVICE_PAUSED)
// )

func controlService(c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(svcName)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

var exampleQuery string = "grant_new_nonce = data.kbridge.grant_new_nonce; data.kbridge.get_principals[get_principals]"

var examplePolicy string = `
package kbridge

default grant_new_nonce = false
grant_new_nonce {
	input.iss == "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo"
}
get_principals[grant] {
	grant := split(input.claims.service.keytab,",")
}
`

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}

	fasttick := time.Tick(10 * time.Second)
	slowtick := time.Tick(60 * time.Second)
	tick := fasttick

	configLoader, err := configloader.NewConfigLoader()
	if err != nil {
		elog.Error(10, err.Error())
		return false, 2
	}

	serverConfig, err := configLoader.ServerConfig()
	if err != nil {
		elog.Error(11, err.Error())
		return false, 2
	}

	zapConfig, err := configLoader.ZapConfig()
	if err != nil {
		elog.Error(12, err.Error())
		return false, 2
	}

	logger, err := zapConfig.Build()
	if err != nil {
		elog.Error(13, err.Error())
		return false, 2
	}

	logger = logger.WithOptions(zap.Hooks(getZapHook()))

	zap.ReplaceGlobals(logger)
	//defer logger.Sync()

	server, err := serverConfig.Build()
	if err != nil {
		elog.Error(1, err.Error())
		return false, 2
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case <-tick:
			// elog.Info(1, "beep")
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				server.Shutdown()
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}

	return
}

func getZapHook() func(zapcore.Entry) error {

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

		case zapcore.DebugLevel:
			//	elog.Info(1, e.Message)
			break

		case zapcore.InfoLevel:
			elog.Info(1, e.Message)
			break

		case zapcore.WarnLevel:
			elog.Warning(1, e.Message)
			break

		// Everthing else
		default:
			elog.Error(1, e.Message)
			break

		}

		return nil
	}
}
