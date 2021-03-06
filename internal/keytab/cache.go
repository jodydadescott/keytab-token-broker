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

	"github.com/jodydadescott/tokens2secrets/internal/timeperiod"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

var (
	principalRegex  = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	defaultLifetime = time.Duration(5) * time.Minute

	passwordCharset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@!"
)

// Config Configuration
//
// Seed: A shared secret that the password for a keytab is generated from
//
// Principals: Zero or more principlas Kerberos principals (or usernames)
//
// TimePeriod: Time Period for Keytab Renewals
type Config struct {
	Keytabs []*Keytab
}

// Cache holds and manages Kerberos Keytabs. Keytabs are generated or
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
type Cache struct {
	closeTimer chan struct{}
	wg         sync.WaitGroup
	ticker     *time.Ticker
	mutex      sync.RWMutex
	internal   map[string]*wrapper
}

type wrapper struct {
	mutex           sync.RWMutex
	nextUpdate      time.Time
	principal, seed string
	keytab          *Keytab
	err             error
	timePeriod      *timeperiod.TimePeriod
}

// Build Returns new instance of Keytabs
func (config *Config) Build() (*Cache, error) {

	zap.L().Debug("Starting")

	t := &Cache{
		closeTimer: make(chan struct{}),
		wg:         sync.WaitGroup{},
		ticker:     time.NewTicker(time.Second),
		internal:   make(map[string]*wrapper),
	}

	err := t.init(config)
	if err != nil {
		return nil, err
	}

	go t.run()
	return t, nil
}

func (t *Cache) init(config *Config) error {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, keytab := range config.Keytabs {
		if len(keytab.Principal) < 3 && len(keytab.Principal) > 254 {
			if len(keytab.Principal) < 3 {
				return fmt.Errorf("Keytab principal %s is to short", keytab.Principal)
			}
			return fmt.Errorf("Keytab principal %s is to long", keytab.Principal)
		}

		if !principalRegex.MatchString(keytab.Principal) {
			return fmt.Errorf("Keytab principal %s is invalid", keytab.Principal)
		}

		if keytab.Seed == "" {
			return fmt.Errorf("Keytab %s is missing required seed", keytab.Principal)
		}

		seed := base32.StdEncoding.EncodeToString([]byte(keytab.Seed))

		lifetime := defaultLifetime
		if keytab.Lifetime > 0 {
			lifetime = keytab.Lifetime
		}

		// Lifetime less then a minute requires to much resources and does not make much sense
		if lifetime < time.Minute {
			return fmt.Errorf(fmt.Sprintf("Keytab %s lifetime is less then one minute. Lifetime must be one minute or greater", keytab.Principal))
		}

		t.internal[keytab.Principal] = &wrapper{
			principal:  keytab.Principal,
			timePeriod: timeperiod.NewPeriod(lifetime),
			seed:       seed,
		}
		zap.L().Debug(fmt.Sprintf("Loaded principal %s", keytab.Principal))
	}

	return nil
}

func (t *Cache) run() {

	zap.L().Debug("Starting")
	t.wg.Add(1)

	timeperiod := timeperiod.NewPeriod(time.Minute)

	next := timeperiod.From(getTime()).Next().Time()

	// We need to run on the top of the time period (or as close as possible). Based on now we
	// set the next run to be the top of the next minute. Technically this means there can be a
	// 59 second delay before the first run based on the time the server is started. We run a
	// ticker every second and check to see if now is equal to or greater then next and if it
	// is we run our jobs and update the next

	for {
		select {
		case <-t.closeTimer:
			t.wg.Done()
			return
		case <-t.ticker.C:
			// This fires every second
			now := getTime()
			if now.Equal(next) || now.After(next) {
				go t.update(next)
				next = timeperiod.From(now).Next().Time()
			}
		}
	}

}

func (t *Cache) update(now time.Time) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	for _, wrapper := range t.internal {
		go wrapper.update(now)
	}
}

func (t *wrapper) update(now time.Time) {

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if now.Equal(t.nextUpdate) || now.After(t.nextUpdate) {

		zap.L().Debug(fmt.Sprintf("Keytab %s ready for new keytab", t.principal))

		nowPeriod := t.timePeriod.From(now)
		now = nowPeriod.Time()

		otp, err := totp.GenerateCodeCustom(t.seed, now, totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsEight,
			Algorithm: otp.AlgorithmSHA512,
		})

		if err != nil {
			zap.L().Error(fmt.Sprintf("Unable to get create keytab %s ; err->%s", t.principal, err.Error()))
			t.err = err
			t.keytab = nil
			return
		}

		hash := sha256.Sum256([]byte(otp + t.seed))

		b := make([]byte, 28)
		for i := range b {
			b[i] = getChar(hash[i])
		}

		password := string(b)

		base64File, err := newKeytab(t.principal, password)

		if err != nil {
			zap.L().Error(fmt.Sprintf("Unable to get create keytab %s ; err->%s", t.principal, err.Error()))
			t.err = err
			t.keytab = nil
			return
		}

		// This allows the admin to verify that different instances of the server are assiging
		// the same password if they have the same seed without revealing the real password
		passwordhash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))[:12]

		t.nextUpdate = nowPeriod.Next().Time()
		t.err = nil
		t.keytab = &Keytab{
			Principal:  "HTTP/" + t.principal,
			Base64File: base64File,
			Exp:        nowPeriod.Time().Unix() + int64(t.timePeriod.Duration.Seconds()),
		}

		zap.L().Debug(fmt.Sprintf("Keytab %s generated; password=%s, exp=%d", t.principal, passwordhash, t.keytab.Exp))
		return

	}

	zap.L().Debug(fmt.Sprintf("Keytab %s NOT ready for new keytab", t.principal))

}

func getChar(b byte) byte {
	bint := int(b)
	charsetlen := len(passwordCharset)
	if int(b) < charsetlen {
		return passwordCharset[bint]
	}
	_, r := bint/charsetlen, bint%charsetlen
	return passwordCharset[r]
}

func int31n(n int, input int64) int32 {
	v := uint32(input >> 31)
	prod := uint64(v) * uint64(n)
	low := uint32(prod)
	if low < uint32(n) {
		thresh := uint32(-n) % uint32(n)
		for low < thresh {
			v = uint32(input >> 31)
			prod = uint64(v) * uint64(n)
			low = uint32(prod)
		}
	}
	return int32(prod >> 32)
}

// GetKeytab Returns Keytab if keytab exist.
func (t *Cache) GetKeytab(principal string) (*Keytab, error) {

	if principal == "" {
		zap.L().Debug("principal is empty")
		return nil, ErrNotFound
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if wrapper, exist := t.internal[principal]; exist {

		wrapper.mutex.RLock()
		defer wrapper.mutex.RUnlock()

		// Export function; returning copy
		if wrapper.keytab == nil {
			if wrapper.err == nil {
				zap.L().Debug(fmt.Sprintf("Keytab %s has not been processed yet", principal))
				return nil, ErrNotReady
			}
			zap.L().Debug(fmt.Sprintf("Keytab %s not generated due to error; err->%s", principal, wrapper.err.Error()))
			return nil, ErrGenFail
		}

		return wrapper.keytab.Copy(), nil
	}

	zap.L().Debug(fmt.Sprintf("Keytab %s does not exist", principal))
	return nil, ErrNotFound
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

	zap.L().Debug(fmt.Sprintf("command->%s, output->%s", logarg, string(cmdOutput.Bytes())))

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

func getTime() time.Time {
	// If running multiple instance the time must be the same so we statically use UTC
	return time.Now().In(time.UTC)
}

// Shutdown shutdown
func (t *Cache) Shutdown() {
	zap.L().Info(fmt.Sprintf("Stopping"))
	close(t.closeTimer)
	t.wg.Wait()
}
