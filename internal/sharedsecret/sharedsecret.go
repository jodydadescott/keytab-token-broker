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

package sharedsecret

import (
	"fmt"
	"time"

	"github.com/jodydadescott/keytab-token-broker/internal/timeperiod"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

var (
	defaultTimePeriod = timeperiod.GetDay()
)

// Config ...
type Config struct {
	Seed       string
	TimePeriod timeperiod.TimePeriod
}

// SharedSecret ...
type SharedSecret struct {
	seed       string
	timePeriod timeperiod.TimePeriod
}

// Build Returns new instance of SharedSecret
func (config *Config) Build() (*SharedSecret, error) {

	if config.Seed == "" {
		return nil, fmt.Errorf("The seed is not set")
	}

	if len(config.Seed) < 20 {
		return nil, fmt.Errorf("The seed is to short")
	}

	timePeriod := defaultTimePeriod

	if config.TimePeriod != "" {
		timePeriod = config.TimePeriod
	}

	return &SharedSecret{
		seed:       config.Seed,
		timePeriod: timePeriod,
	}, nil

}

// GetTimePeriod Returns the time period
func (t *SharedSecret) GetTimePeriod() timeperiod.TimePeriod {
	return t.timePeriod
}

// GetSecret Returns the secret for right now
func (t *SharedSecret) GetSecret() (string, error) {
	return t.getSecretForTime(t.timePeriod.Now(getNow()))
}

// GetNextSecret Returns the secret for the next time
func (t *SharedSecret) GetNextSecret() (string, error) {
	return t.getSecretForTime(t.timePeriod.Next(getNow()))
}

func (t *SharedSecret) getSecretForTime(time time.Time) (string, error) {

	otp, err := totp.GenerateCodeCustom(t.seed, t.timePeriod.Now(time), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA512,
	})
	if err != nil {
		zap.L().Error(fmt.Sprintf("Unable to get secret; err->%s", err))
		return "", err
	}

	return otp, nil
}

func getNow() time.Time {
	// If running multiple instance the time must be the same so we statically use UTC
	return time.Now().In(time.UTC)
}
