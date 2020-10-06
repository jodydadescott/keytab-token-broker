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
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jodydadescott/keytab-token-broker/internal/timeperiod"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

var (
	defaultTimePeriod = timeperiod.GetFiveMinute()
	principalRegex    = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
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
	TimePeriod timeperiod.TimePeriod
}

// Keytabs holds and manages Kerberos Keytabs. Keytabs are generated or
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
type Keytabs struct {
	closeTimer, closeKeymaker chan struct{}
	runKeymaker               chan time.Time
	wg                        sync.WaitGroup
	ticker                    *time.Ticker
	mutex                     sync.RWMutex
	internal                  map[string]*Keytab
	seed                      string
	principals                []string
	timePeriod                timeperiod.TimePeriod
}

// Build Returns new instance of Keytabs
func (config *Config) Build() (*Keytabs, error) {

	var err error

	if config.Seed == "" {
		return nil, fmt.Errorf("Seed is empty")
	}

	timePeriod := defaultTimePeriod

	if config.TimePeriod != "" {
		timePeriod = config.TimePeriod
	}

	t := &Keytabs{
		closeTimer:    make(chan struct{}),
		closeKeymaker: make(chan struct{}),
		runKeymaker:   make(chan time.Time),
		wg:            sync.WaitGroup{},
		ticker:        time.NewTicker(time.Second),
		internal:      make(map[string]*Keytab),
		seed:          base32.StdEncoding.EncodeToString([]byte(config.Seed)),
		timePeriod:    timePeriod,
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
	diff := t.timePeriod.Next(now).Sub(now).Seconds()

	// zap.L().Debug(fmt.Sprintf("Period->Now:%s, PeriodNow:%s, PeriodNext: %s", now, timePeriod.Now(now), timePeriod.Next(now)))

	if diff > 30 {
		zap.L().Debug(fmt.Sprintf("The next Keytab renew is in %f seconds; calling keymmaker now", diff))
		t.runKeymaker <- timePeriod.Now(now)

	} else {
		zap.L().Debug(fmt.Sprintf("The next Keytab renew is in %f seconds; will let the ticker call keymaker when time", diff))
	}

	go func() {
		zap.L().Debug("Starting Ticker")
		t.wg.Add(1)

		next := timePeriod.Next(now)

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
					next = timePeriod.Next(now)
					t.runKeymaker <- timePeriod.Now(now)
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

// func zeroTime(input time.Time) time.Time {
// 	return time.Date(input.Year(), input.Month(), input.Day(), input.Hour(), input.Minute(), 0, 0, time.UTC)
// }

func (t *Keytabs) cacheRefresh(now time.Time) {

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

		// Principal was already verified with from of user@domain. We get the username only
		// for the hash. This means that if the seed is the same for two different instances
		// and so is the username the password will be the same even if
		username := strings.Split(principal, "@")[0]

		password := fmt.Sprintf("%x", sha256.Sum256([]byte(t.seed+username+otp)))[:24]
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
func (t *Keytabs) GetKeytab(principal string) *Keytab {

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if keytab, exist := t.internal[principal]; exist {
		// Export function; returning copy
		return keytab.Clone()
	}

	return nil
}

// Shutdown ...
func (t *Keytabs) Shutdown() {
	close(t.closeTimer)
	close(t.closeKeymaker)
	t.wg.Wait()
}
