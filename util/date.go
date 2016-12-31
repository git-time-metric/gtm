// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"
)

// Now is the func used for system time within gtm
// This allows for manipulating system time during testing
var Now = func() time.Time { return time.Now() }

// DataRange creates predefined date ranges and validates if dates are within the range
type DateRange struct {
	Start time.Time
	End   time.Time
}

// IsSet returns true if the date range has a starting and/or ending date
func (d DateRange) IsSet() bool {
	return !d.Start.IsZero() || !d.End.IsZero()
}

// String returns a date range as a string
func (d DateRange) String() string {
	return fmt.Sprintf("%s - %s", d.Start.Format(time.UnixDate), d.End.Format(time.UnixDate))
}

// Within determines if a date is within the date range
func (r DateRange) Within(t time.Time) bool {
	switch {
	case !r.Start.IsZero() && !r.End.IsZero():
		return t.Equal(r.Start) || t.Equal(r.End) || (t.After(r.Start) && t.Before(r.End))
	case !r.Start.IsZero():
		return t.Equal(r.Start) || t.After(r.Start)
	case !r.End.IsZero():
		return t.Equal(r.End) || t.Before(r.End)
	default:
		return false
	}

}

// AfterNow returns a date range ending n days in the past
func AfterNow(n int) DateRange {
	end := now.New(Now()).EndOfDay().AddDate(0, 0, -n)
	return DateRange{End: end}
}

// TodayRange returns a date range for today
func TodayRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfDay()
	end := now.EndOfDay()

	return DateRange{Start: start, End: end}
}

// YesterdayRange returns a date range for yesterday
func YesterdayRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfDay().AddDate(0, 0, -1)
	end := start.AddDate(0, 0, 1).Add(-time.Nanosecond)

	return DateRange{Start: start, End: end}
}

// ThisWeekRange returns a date range for this week
func ThisWeekRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfWeek()
	end := now.EndOfWeek()

	return DateRange{End: end, Start: start}
}

// LastWeekRange returns a date for last week
func LastWeekRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfWeek().AddDate(0, 0, -7)
	end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)

	return DateRange{End: end, Start: start}
}

// ThisMonthRange returns a date range for this month
func ThisMonthRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfMonth()
	end := now.EndOfMonth()

	return DateRange{End: end, Start: start}
}

// LastMonthRange returns a date range for last month
func LastMonthRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfMonth().AddDate(0, -1, 0)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	return DateRange{End: end, Start: start}
}

// ThisYearRange returns a date range for this year
func ThisYearRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfYear()
	end := now.EndOfYear()

	return DateRange{End: end, Start: start}
}

// LastYearRange returns a date range for last year
func LastYearRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfYear().AddDate(-1, 0, 0)
	end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)

	return DateRange{End: end, Start: start}
}
