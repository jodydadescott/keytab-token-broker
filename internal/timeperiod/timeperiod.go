package timeperiod

import (
	"fmt"
	"time"
)

// Config Configuration
type Config struct {
	// Seconds in period. Must be greater then zero and even
	Seconds int
}

// TimePeriod Provides synchrnous time periods based on epoch seconds and
// period interval seconds
type TimePeriod struct {
	seconds int64
}

// Build Returns new instance from comfig
func (c *Config) Build() (*TimePeriod, error) {

	if c.Seconds <= 0 || c.Seconds%2 != 0 {
		return nil, fmt.Errorf(fmt.Sprintf("TimePeriod Seconds %d is invalid, must be greater then zero and even", c.Seconds))
	}

	return &TimePeriod{
		seconds: int64(c.Seconds),
	}, nil
}

// Now returns current top of period for provided time
func (t *TimePeriod) Now(input time.Time) time.Time {
	nowSeconds := input.Unix()
	timePeriodSeconds := int64(t.seconds)
	_, remainderSeconds := nowSeconds/timePeriodSeconds, nowSeconds%timePeriodSeconds
	return time.Unix(nowSeconds-remainderSeconds, 0).In(input.Location())
}

// Next returns next top of period for provided time
func (t *TimePeriod) Next(input time.Time) time.Time {
	nowSeconds := input.Unix()
	_, remainderSeconds := nowSeconds/t.seconds, nowSeconds%t.seconds
	return time.Unix(nowSeconds-remainderSeconds+t.seconds, 0).In(input.Location())
}

// Prev returns previous top of period for provided time
func (t *TimePeriod) Prev(input time.Time) time.Time {
	nowSeconds := input.Unix()
	_, remainderSeconds := nowSeconds/t.seconds, nowSeconds%t.seconds
	return time.Unix(nowSeconds-remainderSeconds-t.seconds, 0).In(input.Location())
}
