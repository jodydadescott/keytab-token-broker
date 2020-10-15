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

package keytab

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/jodydadescott/tokens2keytabs/internal/timeperiod"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

var (
	principalRegex  = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	defaultLifetime = time.Duration(5) * time.Minute
)

// Config Configuration
//
// Seed: A shared secret that the password for a keytab is generated from
//
// Principals: Zero or more principlas Kerberos principals (or usernames)
//
// TimePeriod: Time Period for Keytab Renewals
type Config struct {
	Seed       string
	Principals []string
	Lifetime   time.Duration
}

// Server holds and manages Kerberos Keytabs. Keytabs are generated or
// regenerated based on user specified intervals using the UNIX
// cron format. When multiple instances of the server are ran the cron
// interval should be configured the same. When keytabs are generated
// or regenerated the password is set based on the Seed value and the
// principal name. Only Principals specified in the Principals slice
// will be generated.
//
// Keytab generation operates indepedentely from Keytab request. When
// a Keytab is requested it will be allocated regardless of the time
// remaining until the next regeneration. For example if a Keytab is
// requested that only has 5 seconds left before regeneration it will
// be returned. This may not be enough time for the client to obtain
// a Kerberos ticket. The renewal period is provided as an expiration
// field in the Keytab. This allows the client to determine of enough
// time remains to obtain the Kerberos ticket and act accordingly by
// for example requesting the Keytab again after the renewal.
//
// When operated in a multi-server configuration it is important that the
// cron renewal period is identical and that the clocks are synchronized.
// Additionally the Seed must match.
//
// The password is derived from the Seed based on the request time. To
// keep the passwords synchronized the requesting time is set based on the
// cron period. When the server is initially started the next and previous
// periods are calculated. If they differ by more then 30 seconds then
// the Keytabs are generated using the previous period. Otherwise they
// will be created when the next period arrives.
type Server struct {
	closeTimer, closeKeymaker chan struct{}
	runKeymaker               chan time.Time
	wg                        sync.WaitGroup
	ticker                    *time.Ticker
	mutex                     sync.RWMutex
	internal                  map[string]*Keytab
	seed                      string
	principals                []string
	timePeriod                *timeperiod.TimePeriod
}

// Build Returns new instance of Keytabs
func (config *Config) Build() (*Server, error) {

	var err error

	if config.Seed == "" {
		return nil, fmt.Errorf("Seed is empty")
	}

	lifetime := defaultLifetime

	if config.Lifetime > 0 {
		lifetime = config.Lifetime
	}

	if err != nil {
		return nil, err
	}

	t := &Server{
		closeTimer:    make(chan struct{}),
		closeKeymaker: make(chan struct{}),
		runKeymaker:   make(chan time.Time),
		wg:            sync.WaitGroup{},
		ticker:        time.NewTicker(time.Second),
		internal:      make(map[string]*Keytab),
		seed:          base32.StdEncoding.EncodeToString([]byte(config.Seed)),
		timePeriod:    timeperiod.NewPeriod(lifetime),
	}

	if err != nil {
		return nil, err
	}

	for _, principal := range config.Principals {
		if len(principal) < 3 && len(principal) > 254 {
			if len(principal) < 3 {
				return nil, fmt.Errorf("Principal %s is to short", principal)
			}
			return nil, fmt.Errorf("Principal %s is to long", principal)
		}

		if !principalRegex.MatchString(principal) {
			return nil, fmt.Errorf("Principal %s is invalid", principal)
		}

		t.principals = append(t.principals, principal)
		zap.L().Debug(fmt.Sprintf("Loaded principal %s", principal))
	}

	// Before starting Ticker check to see if Time Period is almost over. If it is
	// we just start the ticker and let it do its thing. Otherwise we have the keymaker
	// run a loop and then we start the ticker.

	go func() {
		zap.L().Debug("Starting Keytab Maker")
		t.wg.Add(1)
		for {
			select {
			case <-t.closeKeymaker:
				zap.L().Debug("Stopping Keytab Maker")
				t.wg.Done()
				return
			case now := <-t.runKeymaker:
				t.cacheRefresh(now)
			}
		}
	}()

	now := getTime()
	diff := t.timePeriod.From(now).Next().Time().Sub(now).Seconds()

	// zap.L().Debug(fmt.Sprintf("Period->Now:%s, PeriodNow:%s, PeriodNext: %s", now, timePeriod.Now(now), timePeriod.Next(now)))

	if diff > 30 {
		zap.L().Debug(fmt.Sprintf("The next Keytab renew is in %f seconds; calling keymmaker now", diff))
		t.runKeymaker <- t.timePeriod.From(now).Time()

	} else {
		zap.L().Debug(fmt.Sprintf("The next Keytab renew is in %f seconds; will let the ticker call keymaker when time", diff))
	}

	go func() {
		zap.L().Debug("Starting Ticker")
		t.wg.Add(1)

		next := t.timePeriod.From(now).Next().Time()

		for {
			select {
			case <-t.closeTimer:
				zap.L().Debug("Stopping Ticker")
				t.wg.Done()
				return
			case <-t.ticker.C:
				// This fires every second
				now := getTime()
				if now.Equal(next) || now.After(next) {
					next = t.timePeriod.From(now).Next().Time()
					t.runKeymaker <- t.timePeriod.From(now).Time()
				}
			}
		}
	}()

	return t, nil
}

func getTime() time.Time {
	// If running multiple instance the time must be the same so we statically use UTC
	return time.Now().In(time.UTC)
}

func (t *Server) cacheRefresh(now time.Time) {

	zap.L().Debug(fmt.Sprintf("Running cacheRefresh for period:%s", now))

	// The period of 30 seconds is fine because it is less then
	// the smalles possible validity period of 60 seconds
	otp, err := totp.GenerateCodeCustom(t.seed, now, totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA512,
	})
	if err != nil {
		zap.L().Error(fmt.Sprintf("Unable to get secret; err->%s", err))
		return
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	exp := now.Unix()

	for _, principal := range t.principals {

		// Password is composed of seed, principal and OTP to make per keytab
		// unique password. As long as seed and time are the same then passwords will
		// match even when computed independentley

		password := fmt.Sprintf("%x", sha256.Sum256([]byte(t.seed+principal+otp)))[:24]
		// A special character is required but a sha256 operation will not return
		// a special char so for now statically added on a special char
		password = password + `/a`

		// This allows the admin to verify that different instances of the server are assiging
		// the same password if they have the same seed without revealing the real password
		passwordhash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))[:12]

		zap.L().Info(fmt.Sprintf("Principal=%s, Password_Hash=%s", principal, passwordhash))

		base64File, err := newKeytab(principal, password)
		if err != nil {
			zap.L().Error(fmt.Sprintf("Unable to get create keytab for principal %s ; err->%s", principal, err))
			return
		}

		t.internal[principal] = &Keytab{
			Principal:  "HTTP/" + principal,
			Base64File: base64File,
			Exp:        exp,
		}

	}

}

// GetKeytab Returns Keytab if keytab exist.
func (t *Server) GetKeytab(principal string) *Keytab {

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if keytab, exist := t.internal[principal]; exist {
		// Export function; returning copy
		return keytab.Clone()
	}

	return nil
}

func newKeytab(principal, password string) (string, error) {
	if runtime.GOOS == "windows" {
		return windowsNewKeytab(principal, password)
	}
	return unixNewKeytab(principal, password)
}

// Windows Kerberos Implementation (Active Directory) allows for the creation
// of principals that are mapped to a user account. Only one principal may be
// mapped to a user account at a time. Once a keytab is created it will remain
// valid until the principal is removed  or the password is changed or a new
// keytab is created. The windows utility ktpass is used to create the keytabs.
// The ktpass command is executed directly on the host. Therefore this should
// be ran on a Windows system that is a member of the target domain. It must
// also be ran with privileges to allow the creation of keytabs. Generally this
// is a Domain Admin. If running as a service it is necessary that it be
// configured to run as a domain admin or user with the privileges necessary
// to create keytabs.
//
// Information about the ktpass utility is as follows
// Exe: C:\Windows\System32\ktpass
// Documentation: https://docs.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2012-r2-and-2012/cc753771(v=ws.11)
// [/out <FileName>]
// [/princ <PrincipalName>]
// [/mapuser <UserAccount>]
// [/mapop {add|set}] [{-|+}desonly] [/in <FileName>]
// [/pass {Password|*|{-|+}rndpass}]
// [/minpass]
// [/maxpass]
// [/crypto {DES-CBC-CRC|DES-CBC-MD5|RC4-HMAC-NT|AES256-SHA1|AES128-SHA1|All}]
// [/itercount]
// [/ptype {KRB5_NT_PRINCIPAL|KRB5_NT_SRV_INST|KRB5_NT_SRV_HST}]
// [/kvno <KeyVersionNum>]
// [/answer {-|+}]
// [/target]
// [/rawsalt] [{-|+}dumpsalt] [{-|+}setupn] [{-|+}setpass <Password>]  [/?|/h|/help]
//
// Use +DumpSalt to dump MIT Salt to output
//
// Notes about ktpass failure functionality
// Testing on Windows Server 2019 reveals that if the user lacks the
// privileges to create keytabs the ktpass utility does not create the
// keytab but also still exits with 0 and nothing is sent to the stdout
// This was with a service account and stderr was not checked. For this
// reason we will return an auth err if the file does not exist. This
// should be refined in the future.
//
// ktpass -mapUser bob@EXAMPLE.COM -pass ** -mapOp set -crypto AES256-SHA1 -ptype KRB5_NT_PRINCIPAL -princ HTTP/bob@EXAMPLE.COM -out keytab
//
func windowsNewKeytab(principal, password string) (string, error) {

	dir, err := ioutil.TempDir("", "kt")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	filename := dir + `\file.keytab`

	exe := "C:\\Windows\\System32\\ktpass"
	args := []string{}

	args = append(args, "-mapUser")
	args = append(args, principal)
	args = append(args, "-pass")
	args = append(args, password)
	args = append(args, "-mapOp")
	args = append(args, "set")
	args = append(args, "-crypto")
	args = append(args, "AES256-SHA1")
	args = append(args, "-ptype")
	args = append(args, "KRB5_NT_PRINCIPAL")
	args = append(args, "-princ")
	args = append(args, "HTTP/"+principal)
	args = append(args, "-kvno")
	args = append(args, "1")
	args = append(args, "-out")
	args = append(args, filename)

	logarg := exe
	for _, arg := range args {
		logarg = logarg + " " + arg
	}

	//zap.L().Debug(fmt.Sprintf("command->%s", logarg))

	cmd := exec.Command(exe, args...)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err != nil {
		zap.L().Error(fmt.Sprintf("exec.Command(%s, %s)", exe, args))
		return "", err
	}

	zap.L().Info(fmt.Sprintf("command->%s, output->%s", logarg, string(cmdOutput.Bytes())))

	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

func unixNewKeytab(principal, password string) (string, error) {
	return "this is not a valid keytab, it is fake", nil
}

// Shutdown ...
func (t *Server) Shutdown() {
	close(t.closeTimer)
	close(t.closeKeymaker)
	t.wg.Wait()
}
