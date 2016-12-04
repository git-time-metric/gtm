package util

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"
)

// Now is the func used for system time within gtm
// This allows for manipulating system time during testing
var Now = func() time.Time { return time.Now() }

type DateRange struct {
	Start time.Time
	End   time.Time
}

func (d DateRange) IsSet() bool {
	return !d.Start.IsZero() || !d.End.IsZero()
}

func (d DateRange) String() string {
	return fmt.Sprintf("%s - %s", d.Start.Format(time.UnixDate), d.End.Format(time.UnixDate))
}

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

func TodayRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfDay()
	end := now.EndOfDay()

	return DateRange{Start: start, End: end}
}

func YesterdayRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfDay().AddDate(0, 0, -1)
	end := start.AddDate(0, 0, 1).Add(-time.Nanosecond)

	return DateRange{Start: start, End: end}
}

func ThisWeekRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfWeek()
	end := now.EndOfWeek()

	return DateRange{End: end, Start: start}
}

func LastWeekRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfWeek().AddDate(0, 0, -7)
	end := start.AddDate(0, 0, 7).Add(-time.Nanosecond)

	return DateRange{End: end, Start: start}
}

func ThisMonthRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfMonth()
	end := now.EndOfMonth()

	return DateRange{End: end, Start: start}
}

func LastMonthRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfMonth().AddDate(0, -1, 0)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)

	return DateRange{End: end, Start: start}
}

func ThisYearRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfYear()
	end := now.EndOfYear()

	return DateRange{End: end, Start: start}
}

func LastYearRange() DateRange {
	now := now.New(Now())

	start := now.BeginningOfYear().AddDate(-1, 0, 0)
	end := start.AddDate(1, 0, 0).Add(-time.Nanosecond)

	return DateRange{End: end, Start: start}
}
