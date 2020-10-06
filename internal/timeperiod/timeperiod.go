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
	"strings"
	"time"
)

// type TimePeriod struct {
// 	ptime time.Time
// }

// type Config struct {
// 	ptime time.Time
// }

// TimePeriod ,,,
type TimePeriod string

const (
	// OneMinute ...
	OneMinute TimePeriod = "OneMinute"
	// FiveMinute ...
	FiveMinute = "FiveMinute"
	// QuarterHour ...
	QuarterHour = "QuarterHour"
	// HalfHour ...
	HalfHour = "HalfHour"
	// Hour ...
	Hour = "Hour"
	// QuarterDay ...
	QuarterDay = "QuarterDay"
	// HalfDay ...
	HalfDay = "HalfDay"
	// Day ...
	Day = "Day"
)

// GetOneMinute ...
func GetOneMinute() TimePeriod {
	return OneMinute
}

// GetFiveMinute ...
func GetFiveMinute() TimePeriod {
	return FiveMinute
}

// GetQuarterHour ...
func GetQuarterHour() TimePeriod {
	return QuarterHour
}

// GetHalfHour ...
func GetHalfHour() TimePeriod {
	return HalfHour
}

// GetHour ...
func GetHour() TimePeriod {
	return Hour
}

// GetQuarterDay ...
func GetQuarterDay() TimePeriod {
	return QuarterDay
}

// GetHalfDay ...
func GetHalfDay() TimePeriod {
	return HalfDay
}

// GetDay ...
func GetDay() TimePeriod {
	return Day
}

// GetTimePeriodFromString Return time period from string
func GetTimePeriodFromString(timeperiod string) (TimePeriod, error) {

	switch strings.ToLower(timeperiod) {

	case "oneminute":
		return OneMinute, nil

	case "fiveminute":
		return FiveMinute, nil

	case "quarterhour":
		return QuarterHour, nil

	case "halfhour":
		return HalfHour, nil

	case "hour":
		return Hour, nil

	case "quarterday":
		return QuarterDay, nil

	case "halfday":
		return HalfDay, nil

	case "day":
		return Day, nil

	}

	return "", fmt.Errorf(fmt.Sprintf("%s is not a known time period", timeperiod))
}

// Now TimePeriod happening now now
func (t TimePeriod) Now(now time.Time) time.Time {

	switch t {

	case OneMinute:
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())

	case FiveMinute:
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), fiveMinute(now.Minute()), 0, 0, now.Location())

	case QuarterHour:
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), quarterHour(now.Minute()), 0, 0, now.Location())

	case HalfHour:
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), halfHour(now.Minute()), 0, 0, now.Location())

	case Hour:
		return now.Truncate(time.Hour)

	case QuarterDay:
		return time.Date(now.Year(), now.Month(), now.Day(), quarterDay(now.Hour()), 0, 0, 0, now.Location())

	case HalfDay:
		return time.Date(now.Year(), now.Month(), now.Day(), halfDay(now.Hour()), 0, 0, 0, now.Location())

	case Day:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	}

	panic("this should never happen")
}

// Next TimePeriod happening after now
func (t TimePeriod) Next(now time.Time) time.Time {

	switch t {

	case OneMinute:
		return t.Now(now.Add(time.Duration(time.Minute)))

	case FiveMinute:
		return t.Now(now.Add(time.Duration(time.Minute * 5)))

	case QuarterHour:
		return t.Now(now.Add(time.Duration(time.Minute * 15)))

	case HalfHour:
		return t.Now(now.Add(time.Duration(time.Minute * 30)))

	case Hour:
		return t.Now(now.Add(time.Duration(time.Hour)))

	case QuarterDay:
		return t.Now(now.Add(time.Duration(time.Hour * 6)))

	case HalfDay:
		return t.Now(now.Add(time.Duration(time.Hour * 12)))

	case Day:
		return t.Now(now.Add(time.Duration(time.Hour * 24)))

	}

	panic("this should never happen")

}

// Prev TimePeriod happened previous to now
func (t TimePeriod) Prev(now time.Time) time.Time {

	switch t {

	case OneMinute:
		return t.Now(now.Add(time.Duration(time.Minute) * -1))

	case FiveMinute:
		return t.Now(now.Add(time.Duration(time.Minute * -5)))

	case QuarterHour:
		return t.Now(now.Add(time.Duration(time.Minute * -15)))

	case HalfHour:
		return t.Now(now.Add(time.Duration(time.Minute * -30)))

	case Hour:
		return t.Now(now.Add(time.Duration(time.Hour)))

	case QuarterDay:
		return t.Now(now.Add(time.Duration(time.Hour * -6)))

	case HalfDay:
		return t.Now(now.Add(time.Duration(time.Hour * -12)))

	case Day:
		return t.Now(now.Add(time.Duration(time.Hour * -24)))

	}

	panic("this should never happen")

}

func halfDay(input int) int {

	if input > 23 {
		panic("unexpected")
	}

	if input >= 12 {
		return 12
	}

	return 0
}

func quarterDay(input int) int {

	if input > 23 {
		panic("this should never happen")
	}

	if input >= 18 {
		return 18
	}

	if input >= 12 {
		return 12
	}

	if input >= 6 {
		return 6
	}

	return 0
}

func halfHour(input int) int {

	if input > 59 {
		panic("this should never happen")
	}

	if input >= 30 {
		return 30
	}

	return 0
}

func quarterHour(input int) int {

	if input > 59 {
		panic("this should never happen")
	}

	if input >= 45 {
		return 45
	}

	if input >= 30 {
		return 30
	}

	if input >= 15 {
		return 15
	}

	return 0
}

func fiveMinute(input int) int {

	if input > 59 {
		panic("this should never happen")
	}

	if input >= 55 {
		return 55
	}

	if input >= 50 {
		return 50
	}

	if input >= 45 {
		return 45
	}

	if input >= 40 {
		return 40
	}
	if input >= 35 {
		return 35
	}
	if input >= 30 {
		return 30
	}
	if input >= 25 {
		return 25
	}
	if input >= 20 {
		return 20
	}
	if input >= 15 {
		return 15
	}
	if input >= 10 {
		return 10
	}
	if input >= 5 {
		return 5
	}

	return 0

}
