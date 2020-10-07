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

package timeperiod

import (
	"fmt"
	"testing"
	"time"
)

func Test1(t *testing.T) {

	{
		c := &Config{
			Seconds: 3601,
		}

		_, err := c.Build()

		if err == nil {
			t.Fatalf("not expected")
		}
	}

	{
		c := &Config{
			Seconds: 0,
		}

		_, err := c.Build()

		if err == nil {
			t.Fatalf("not expected")
		}
	}

	{
		c := &Config{
			Seconds: -1,
		}

		_, err := c.Build()

		if err == nil {
			t.Fatalf("not expected")
		}
	}

	{
		c := &Config{
			Seconds: 128,
		}

		_, err := c.Build()

		if err != nil {
			t.Fatalf("not expected")
		}
	}

}

func Test2(t *testing.T) {

	now := time.Date(2020, 3, 12, 14, 10, 0, 0, time.UTC)

	c := &Config{
		Seconds: 3600,
	}

	timePeriod, _ := c.Build()

	nowPeriod := timePeriod.Now(now)
	nextPeriod := timePeriod.Next(now)
	prevPeriod := timePeriod.Prev(now)

	if fmt.Sprintf("%s", nowPeriod) != "2020-03-12 14:00:00 +0000 UTC" {
		t.Fatalf("nowPeriod is incorrect")
	}

	if fmt.Sprintf("%s", prevPeriod) != "2020-03-12 13:00:00 +0000 UTC" {
		t.Fatalf("prevPeriod is incorrect")
	}

	if fmt.Sprintf("%s", nextPeriod) != "2020-03-12 15:00:00 +0000 UTC" {
		t.Fatalf("nextPeriod is incorrect")
	}

}
