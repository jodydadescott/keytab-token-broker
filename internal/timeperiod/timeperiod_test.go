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

// TimePeriod
// OneMinute
// FiveMinute
// QuarterHour
// HalfHour
// Hour
// QuarterDay
// HalfDay
// Day

// now := time.Now()
// oneMinutePeriod := GetOneMinute().TopOfPeriod(now)
// fiveMinutePeriod := GetFiveMinute().TopOfPeriod(now)
// quaterHourPeriod := GetQuarterHour().TopOfPeriod(now)
// halfHourPeriod := GetHalfHour().TopOfPeriod(now)
// hourPeriod := GetHour().TopOfPeriod(now)
// quarterDayPeriod := GetQuarterDay().TopOfPeriod(now)
// halfDayPeriod := GetHalfDay().TopOfPeriod(now)
// dayPeriod := GetDay().TopOfPeriod(now)
// fmt.Println("oneMinutePeriod: ", oneMinutePeriod)
// fmt.Println("fiveMinutePeriod: ", fiveMinutePeriod)
// fmt.Println("quaterHourPeriod: ", quaterHourPeriod)
// fmt.Println("halfHourPeriod: ", halfHourPeriod)
// fmt.Println("hourPeriod: ", hourPeriod)
// fmt.Println("quarterDayPeriod: ", quarterDayPeriod)
// fmt.Println("halfDayPeriod: ", halfDayPeriod)
// fmt.Println("dayPeriod: ", dayPeriod)

func Test1(t *testing.T) {

	testTime := time.Date(2020, 10, 4, 15, 55, 0, 0, time.UTC)

	oneMinutePeriod := fmt.Sprintf("%s", GetOneMinute().Now(testTime))
	fiveMinutePeriod := fmt.Sprintf("%s", GetFiveMinute().Now(testTime))
	quaterHourPeriod := fmt.Sprintf("%s", GetQuarterHour().Now(testTime))
	halfHourPeriod := fmt.Sprintf("%s", GetHalfHour().Now(testTime))
	hourPeriod := fmt.Sprintf("%s", GetHour().Now(testTime))
	quarterDayPeriod := fmt.Sprintf("%s", GetQuarterDay().Now(testTime))
	halfDayPeriod := fmt.Sprintf("%s", GetHalfDay().Now(testTime))
	dayPeriod := fmt.Sprintf("%s", GetDay().Now(testTime))

	if oneMinutePeriod != "2020-10-04 15:55:00 +0000 UTC" {
		t.Fatalf("oneMinutePeriod is incorrect")
	}

	if fiveMinutePeriod != "2020-10-04 15:55:00 +0000 UTC" {
		t.Fatalf("fiveMinutePeriod is incorrect")
	}

	if quaterHourPeriod != "2020-10-04 15:45:00 +0000 UTC" {
		t.Fatalf("quaterHourPeriod is incorrect")
	}

	if halfHourPeriod != "2020-10-04 15:30:00 +0000 UTC" {
		t.Fatalf("halfHourPeriod is incorrect")
	}

	if hourPeriod != "2020-10-04 15:00:00 +0000 UTC" {
		t.Fatalf("hourPeriod is incorrect")
	}

	if quarterDayPeriod != "2020-10-04 12:00:00 +0000 UTC" {
		t.Fatalf("quarterDayPeriod is incorrect")
	}

	if halfDayPeriod != "2020-10-04 12:00:00 +0000 UTC" {
		t.Fatalf("halfDayPeriod is incorrect")
	}

	if dayPeriod != "2020-10-04 00:00:00 +0000 UTC" {
		t.Fatalf("dayPeriod is incorrect")
	}

	oneMinutePeriodNext := fmt.Sprintf("%s", GetOneMinute().Next(testTime))
	fiveMinutePeriodNext := fmt.Sprintf("%s", GetFiveMinute().Next(testTime))
	quaterHourPeriodNext := fmt.Sprintf("%s", GetQuarterHour().Next(testTime))
	halfHourPeriodNext := fmt.Sprintf("%s", GetHalfHour().Next(testTime))
	hourPeriodNext := fmt.Sprintf("%s", GetHour().Next(testTime))
	quarterDayPeriodNext := fmt.Sprintf("%s", GetQuarterDay().Next(testTime))
	halfDayPeriodNext := fmt.Sprintf("%s", GetHalfDay().Next(testTime))
	dayPeriodNext := fmt.Sprintf("%s", GetDay().Next(testTime))

	if oneMinutePeriodNext != "2020-10-04 15:56:00 +0000 UTC" {
		t.Fatalf("oneMinutePeriod is incorrect")
	}

	if fiveMinutePeriodNext != "2020-10-04 16:00:00 +0000 UTC" {
		t.Fatalf("fiveMinutePeriod is incorrect")
	}

	if quaterHourPeriodNext != "2020-10-04 16:00:00 +0000 UTC" {
		t.Fatalf("quaterHourPeriod is incorrect")
	}

	if halfHourPeriodNext != "2020-10-04 16:00:00 +0000 UTC" {
		t.Fatalf("halfHourPeriod is incorrect")
	}

	if hourPeriodNext != "2020-10-04 16:00:00 +0000 UTC" {
		t.Fatalf("hourPeriod is incorrect")
	}

	if quarterDayPeriodNext != "2020-10-04 18:00:00 +0000 UTC" {
		t.Fatalf("quarterDayPeriod is incorrect")
	}

	if halfDayPeriodNext != "2020-10-05 00:00:00 +0000 UTC" {
		t.Fatalf("halfDayPeriod is incorrect")
	}

	if dayPeriodNext != "2020-10-05 00:00:00 +0000 UTC" {
		t.Fatalf("dayPeriod is incorrect")
	}

	oneMinutePeriodPrev := fmt.Sprintf("%s", GetOneMinute().Prev(testTime))
	fiveMinutePeriodPrev := fmt.Sprintf("%s", GetFiveMinute().Prev(testTime))
	quaterHourPeriodPrev := fmt.Sprintf("%s", GetQuarterHour().Prev(testTime))
	halfHourPeriodPrev := fmt.Sprintf("%s", GetHalfHour().Prev(testTime))
	hourPeriodPrev := fmt.Sprintf("%s", GetHour().Prev(testTime))
	quarterDayPeriodPrev := fmt.Sprintf("%s", GetQuarterDay().Prev(testTime))
	halfDayPeriodPrev := fmt.Sprintf("%s", GetHalfDay().Prev(testTime))
	dayPeriodPrev := fmt.Sprintf("%s", GetDay().Prev(testTime))

	if oneMinutePeriodPrev != "2020-10-04 15:54:00 +0000 UTC" {
		t.Fatalf("oneMinutePeriod is incorrect")
	}

	if fiveMinutePeriodPrev != "2020-10-04 15:50:00 +0000 UTC" {
		t.Fatalf("fiveMinutePeriod is incorrect")
	}

	if quaterHourPeriodPrev != "2020-10-04 15:30:00 +0000 UTC" {
		t.Fatalf("quaterHourPeriod is incorrect")
	}

	if halfHourPeriodPrev != "2020-10-04 15:00:00 +0000 UTC" {
		t.Fatalf("halfHourPeriod is incorrect")
	}

	if hourPeriodPrev != "2020-10-04 16:00:00 +0000 UTC" {
		t.Fatalf("hourPeriod is incorrect")
	}

	if quarterDayPeriodPrev != "2020-10-04 06:00:00 +0000 UTC" {
		t.Fatalf("quarterDayPeriod is incorrect")
	}

	if halfDayPeriodPrev != "2020-10-04 00:00:00 +0000 UTC" {
		t.Fatalf("halfDayPeriod is incorrect")
	}

	if dayPeriodPrev != "2020-10-03 00:00:00 +0000 UTC" {
		t.Fatalf("dayPeriod is incorrect")
	}

}
